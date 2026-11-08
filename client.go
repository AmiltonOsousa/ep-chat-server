package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:4444")
	if err != nil {
		fmt.Println("não consegui conectar:", err)
		return
	}
	defer conn.Close()

	teclado := bufio.NewReader(os.Stdin)
	servidor := bufio.NewReader(conn)

	// o usuário escolhe o apelido antes de entrar
	fmt.Print("Digite seu apelido: ")
	nick, _ := teclado.ReadString('\n')
	nick = strings.TrimSpace(nick)
	fmt.Fprintln(conn, "NICK "+nick)

	// goroutine que imprime o que chega do servidor
	fim := make(chan bool, 1)
	go func() {
		for {
			linha, err := servidor.ReadString('\n')
			if err != nil {
				break
			}
			fmt.Print(linha)
		}
		fmt.Println("Conexão encerrada.")
		fim <- true
		os.Exit(0)
	}()

	// lê o teclado e manda para o servidor
	for {
		texto, err := teclado.ReadString('\n')
		if err != nil {
			break
		}
		texto = strings.TrimSpace(texto)
		if texto == "" {
			continue
		}

		if texto == "\\exit" {
			fmt.Fprintln(conn, "EXIT")
			break
		} else if strings.HasPrefix(texto, "\\changenick ") {
			novo := strings.TrimSpace(strings.TrimPrefix(texto, "\\changenick "))
			fmt.Fprintln(conn, "CHANGENICK "+novo)
		} else if strings.HasPrefix(texto, "\\msg @") {
			// \\msg @apelido mensagem
			resto := strings.TrimPrefix(texto, "\\msg @")
			partes := strings.SplitN(resto, " ", 2)
			if len(partes) < 2 {
				fmt.Println("uso: \\msg @<apelido> <mensagem>")
				continue
			}
			fmt.Fprintln(conn, "PRIV "+partes[0]+" "+partes[1])
		} else if strings.HasPrefix(texto, "\\msg ") {
			fmt.Fprintln(conn, "MSG "+strings.TrimPrefix(texto, "\\msg "))
		} else {
			fmt.Println("comando inválido. Use \\msg, \\changenick ou \\exit")
		}
	}
	fmt.Fprintln(conn, "EXIT")
	// espera o servidor fechar a conexão
	<-fim
}
