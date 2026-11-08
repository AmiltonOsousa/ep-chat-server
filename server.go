package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
	"sync"
)

type Client struct {
	conn net.Conn
	nick string
	bot  bool
	ch   chan string // canal exclusivo do cliente
}

var clients = make(map[string]*Client)
var mutex sync.Mutex

// manda a mensagem para todos os usuários (bots não recebem)
func broadcast(msg string) {
	fmt.Println(msg)
	mutex.Lock()
	for _, c := range clients {
		if !c.bot {
			c.ch <- msg
		}
	}
	mutex.Unlock()
}

// manda mensagem privada para um cliente só
func private(from *Client, to string, msg string) {
	// o lock fica preso até enviar, assim o destinatário não fecha o canal no meio
	mutex.Lock()
	defer mutex.Unlock()
	dest, ok := clients[to]
	if !ok {
		from.ch <- "Usuário @" + to + " não existe"
		return
	}
	fmt.Println("@" + from.nick + " disse em privado para @" + to + ": " + msg)
	if dest.bot {
		// o bot recebe num formato fácil de ler: FROM apelido mensagem
		dest.ch <- "FROM " + from.nick + " " + msg
	} else {
		dest.ch <- "@" + from.nick + " disse em privado: " + msg
	}
}

// troca o apelido de um cliente
func changeNick(c *Client, novo string) {
	mutex.Lock()
	if _, existe := clients[novo]; existe || novo == "" {
		mutex.Unlock()
		c.ch <- "Apelido inválido ou já em uso"
		return
	}
	antigo := c.nick
	delete(clients, antigo)
	c.nick = novo
	clients[novo] = c
	mutex.Unlock()
	broadcast("Usuário @" + antigo + " agora é @" + novo)
}

// goroutine que escreve no socket tudo que chega no canal do cliente
func writer(c *Client) {
	for msg := range c.ch {
		fmt.Fprintln(c.conn, msg)
	}
}

func handle(conn net.Conn) {
	defer conn.Close()
	reader := bufio.NewReader(conn)

	// primeira linha: NICK apelido (usuário) ou BOT apelido (bot)
	line, err := reader.ReadString('\n')
	if err != nil {
		return
	}
	line = strings.TrimSpace(line)
	isBot := strings.HasPrefix(line, "BOT ")
	nick := line
	nick = strings.TrimPrefix(nick, "NICK ")
	nick = strings.TrimPrefix(nick, "BOT ")
	nick = strings.TrimPrefix(strings.TrimSpace(nick), "@")

	c := &Client{conn: conn, nick: nick, bot: isBot, ch: make(chan string, 100)}

	mutex.Lock()
	_, existe := clients[nick]
	if nick == "" || existe {
		mutex.Unlock()
		fmt.Fprintln(conn, "Apelido inválido ou já em uso")
		return
	}
	clients[nick] = c
	mutex.Unlock()
	go writer(c)

	if isBot {
		broadcast("Bot @" + nick + " acabou de entrar")
	} else {
		broadcast("Usuário @" + nick + " acabou de entrar")
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "MSG ") {
			broadcast("@" + c.nick + " disse: " + strings.TrimPrefix(line, "MSG "))
		} else if strings.HasPrefix(line, "PRIV ") {
			partes := strings.SplitN(strings.TrimPrefix(line, "PRIV "), " ", 2)
			if len(partes) == 2 {
				private(c, partes[0], partes[1])
			}
		} else if strings.HasPrefix(line, "CHANGENICK ") {
			changeNick(c, strings.TrimPrefix(strings.TrimSpace(strings.TrimPrefix(line, "CHANGENICK ")), "@"))
		} else if line == "EXIT" {
			break
		}
	}

	// saiu
	mutex.Lock()
	delete(clients, c.nick)
	close(c.ch)
	mutex.Unlock()
	broadcast("Usuário @" + c.nick + " saiu")
}

func main() {
	ln, err := net.Listen("tcp", "localhost:4444")
	if err != nil {
		fmt.Println("erro:", err)
		return
	}
	fmt.Println("Servidor rodando na porta 4444")
	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handle(conn)
	}
}
