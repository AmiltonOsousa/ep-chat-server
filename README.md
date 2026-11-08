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
