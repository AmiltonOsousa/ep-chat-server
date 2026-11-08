package main

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

// inverte o texto (usando runes para não quebrar acentos)
func inverter(texto string) string {
	letras := []rune(texto)
	for i, j := 0, len(letras)-1; i < j; i, j = i+1, j-1 {
		letras[i], letras[j] = letras[j], letras[i]
	}
	return string(letras)
}

func main() {
	conn, err := net.Dial("tcp", "localhost:4444")
	if err != nil {
		fmt.Println("não consegui conectar:", err)
		return
	}
	defer conn.Close()

	// avisa o servidor que é um bot
	fmt.Fprintln(conn, "BOT Inversor")
	fmt.Println("Bot @Inversor conectado. Mande \\msg @Inversor <texto> pelo cliente.")

	servidor := bufio.NewReader(conn)
	for {
		linha, err := servidor.ReadString('\n')
		if err != nil {
			break
		}
		linha = strings.TrimSpace(linha)

		// só reage a mensagens no formato: FROM apelido mensagem
		// o resto (ex: aviso de apelido em uso) só é mostrado no console
		if !strings.HasPrefix(linha, "FROM ") {
			fmt.Println("Servidor: " + linha)
			continue
		}
		partes := strings.SplitN(strings.TrimPrefix(linha, "FROM "), " ", 2)
		if len(partes) < 2 {
			continue
		}
		quem := partes[0]
		texto := partes[1]
		resposta := inverter(texto)

		fmt.Println("Mensagem de: @" + quem)
		fmt.Println("Recebi: " + texto)
		fmt.Println("Resposta: " + resposta)

		fmt.Fprintln(conn, "PRIV "+quem+" "+resposta)
	}
	fmt.Println("Conexão encerrada.")
}
