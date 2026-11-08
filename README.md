# EP03 - Chat server em Go

Três programas separados, todos em `localhost:4444`:

- `server.go`: servidor do chat
- `client.go`: cliente para os usuários
- `bot.go`: bot que inverte o texto recebido

## Como rodar

Em terminais diferentes:

```
go run server.go
go run client.go
go run bot.go
```

## Comandos do cliente

| Comando | Efeito |
|---|---|
| `\changenick <apelido>` | troca o apelido |
| `\msg <mensagem>` | mensagem pública |
| `\msg @<apelido> <mensagem>` | mensagem privada |
| `\exit` | sai do chat |

Para usar o bot, mande uma mensagem privada: `\msg @Inversor ola`

Texto digitado sem comando também é enviado como mensagem pública.

## Protocolo (cliente/bot -> servidor)

| Linha | Significado |
|---|---|
| `NICK <apelido>` | primeira linha de um usuário |
| `BOT <apelido>` | primeira linha de um bot (servidor anuncia como `Bot @apelido`) |
| `MSG <texto>` | mensagem pública |
| `PRIV <apelido> <texto>` | mensagem privada |
| `CHANGENICK <apelido>` | troca de apelido |
| `EXIT` | sair |

O servidor manda para os usuários linhas prontas para imprimir. Para os bots ele manda
`FROM <apelido> <texto>`, que é mais fácil de separar.

## Organização do servidor

- cada cliente tem um canal exclusivo (`ch`) e uma goroutine `writer` que escreve no socket
- `broadcast`: mensagens públicas e avisos (entrada, saída, troca de apelido), só para usuários
- `tratarPrivadas`: mensagens privadas para usuários
- `tratarBots`: mensagens privadas para bots
- `tratarComandos`: troca de apelido
- o mapa de clientes é protegido por um `sync.Mutex`

## Apresentação

Slides com a explicação da implementação: `EP03-chat-server.pptx`.
