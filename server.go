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
	ch   chan string   // canal exclusivo do cliente
	done chan struct{} // fecha quando o cliente sai
}

type Privada struct {
	de    *Client
	para  string
	texto string
}

type Comando struct {
	cliente *Client
	novo    string
}

var clients = make(map[string]*Client)
var mutex sync.Mutex

// canais de cada responsabilidade do servidor
var publicas = make(chan string, 100)
var privadas = make(chan Privada, 100)
var paraBots = make(chan Privada, 100)
var comandos = make(chan Comando, 100)

// envia para o canal do cliente sem travar caso ele esteja lento ou tenha saído
func enviar(c *Client, msg string) {
	select {
	case c.ch <- msg:
	default:
	}
}

// goroutine 1: broadcast das mensagens públicas (bots não recebem)
func broadcast() {
	for msg := range publicas {
		fmt.Println(msg)
		mutex.Lock()
		for _, c := range clients {
			if !c.bot {
				enviar(c, msg)
			}
		}
		mutex.Unlock()
	}
}

// goroutine 2: mensagens privadas para usuários
func tratarPrivadas() {
	for p := range privadas {
		mutex.Lock()
		dest, ok := clients[p.para]
		mutex.Unlock()
		if !ok {
			enviar(p.de, "Usuário @"+p.para+" não existe")
			continue
		}
		// mesmo formato do enunciado, com o destinatário na frente para saber para quem foi
		fmt.Println("[para @" + p.para + "] @" + p.de.nome() + " disse em privado: " + p.texto)
		if dest.bot {
			paraBots <- p // quem trata bot é outra goroutine
		} else {
			enviar(dest, "@"+p.de.nome()+" disse em privado: "+p.texto)
		}
	}
}

// goroutine 3: mensagens privadas para bots
func tratarBots() {
	for p := range paraBots {
		mutex.Lock()
		bot, ok := clients[p.para]
		mutex.Unlock()
		if ok {
			// o bot recebe num formato fácil de ler: FROM apelido mensagem
			enviar(bot, "FROM "+p.de.nome()+" "+p.texto)
		}
	}
}

// goroutine 4: comandos, por enquanto só a troca de apelido
func tratarComandos() {
	for cmd := range comandos {
		c := cmd.cliente
		mutex.Lock()
		_, existe := clients[cmd.novo]
		if existe || cmd.novo == "" || strings.Contains(cmd.novo, " ") {
			mutex.Unlock()
			enviar(c, "Apelido inválido ou já em uso")
			continue
		}
		antigo := c.nick
		delete(clients, antigo)
		c.nick = cmd.novo
		clients[cmd.novo] = c
		mutex.Unlock()
		publicas <- "Usuário @" + antigo + " agora é @" + cmd.novo
	}
}

// goroutine de cada cliente: escreve no socket tudo que chega no canal dele
func writer(c *Client) {
	for {
		select {
		case msg := <-c.ch:
			fmt.Fprintln(c.conn, msg)
		case <-c.done:
			return
		}
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
	nick := strings.TrimPrefix(line, "NICK ")
	nick = strings.TrimPrefix(nick, "BOT ")
	nick = strings.TrimPrefix(strings.TrimSpace(nick), "@")

	c := &Client{conn: conn, nick: nick, bot: isBot,
		ch: make(chan string, 100), done: make(chan struct{})}

	mutex.Lock()
	_, existe := clients[nick]
	if nick == "" || strings.Contains(nick, " ") || existe {
		mutex.Unlock()
		fmt.Fprintln(conn, "Apelido inválido ou já em uso")
		return
	}
	clients[nick] = c
	mutex.Unlock()
	go writer(c)

	if isBot {
		publicas <- "Bot @" + nick + " acabou de entrar"
	} else {
		publicas <- "Usuário @" + nick + " acabou de entrar"
	}

	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "MSG ") {
			publicas <- "@" + c.nome() + " disse: " + strings.TrimPrefix(line, "MSG ")
		} else if strings.HasPrefix(line, "PRIV ") {
			partes := strings.SplitN(strings.TrimPrefix(line, "PRIV "), " ", 2)
			if len(partes) == 2 {
				privadas <- Privada{c, partes[0], partes[1]}
			}
		} else if strings.HasPrefix(line, "CHANGENICK ") {
			novo := strings.TrimPrefix(strings.TrimSpace(strings.TrimPrefix(line, "CHANGENICK ")), "@")
			comandos <- Comando{c, novo}
		} else if line == "EXIT" {
			break
		}
	}

	// saiu
	mutex.Lock()
	nomeFinal := c.nick
	delete(clients, nomeFinal)
	mutex.Unlock()
	close(c.done)
	publicas <- "Usuário @" + nomeFinal + " saiu"
}

func main() {
	ln, err := net.Listen("tcp", "localhost:4444")
	if err != nil {
		fmt.Println("erro:", err)
		return
	}
	fmt.Println("Servidor rodando na porta 4444")

	go broadcast()
	go tratarPrivadas()
	go tratarBots()
	go tratarComandos()

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handle(conn)
	}
}

// lê o apelido com lock, já que a goroutine de comandos pode trocá-lo
func (c *Client) nome() string {
	mutex.Lock()
	defer mutex.Unlock()
	return c.nick
}
