# AGENTS.md

Orientações para agentes de código que trabalham neste repositório.

## Visão geral

Servidor de xadrez multiplayer sobre TCP em Go (projeto da disciplina de Redes de Computadores). O progresso e o escopo planejado ficam em `ROADMAP.md`, em português: marque os itens concluídos ao terminar uma tarefa.

## Ambiente e comandos

O ambiente vem do `flake.nix` (Go 1.27, gopls, gotools, netcat, ent) via direnv (`use flake`) ou `nix develop`. As dependências estão em `vendor/`, então build e testes usam o vendor por padrão.

- Build: `go build ./...`
- Rodar o servidor (escuta em `:9000`): `go run ./cmd/backend`
- Testes: `go test ./internal/...` / um teste: `go test ./internal/domain/chess -run NomeDoTeste`
- Regenerar o código do ent após mudar `ent/schema/`: `go generate ./ent` (todo o resto de `ent/` é gerado; não editar à mão)
- Testar o protocolo manualmente: `nc localhost 9000` e enviar uma linha JSON, ex.: `{"id":"1","type":"login","payload":{"name":"a","password":"b"}}`

O código usa métodos genéricos (ex.: `func (c *TCPClient) SendRequest[T any](...)`), recurso de versões recentes do Go; não é erro de sintaxe.

## Arquitetura

Camadas em `internal/`, só estas três:

- `domain/` — regras e interfaces de serviço (`chess.GameService`, `auth.LoginService`, `users.UserService`). O pacote `chess` mantém o tabuleiro como matriz `[8][8]*piece` (`board.pieces[x][y]`), com tipos internos não exportados; `Game` controla turno e delega o lance ao `board`.
- `infra/` — implementações dos serviços (persistência via ent, hash Argon2id em `password.go`, estado em memória) e o transporte em `infra/tcp/`.
- `application/messages/` — payloads JSON de requisição/resposta compartilhados entre servidor e cliente. Payloads novos vão aqui.

`cmd/backend/main.go` monta os serviços, cria os handlers (`cmd/backend/handlers/`) e registra cada tipo de mensagem no `HandlerManager`.

### Protocolo TCP (`internal/infra/tcp/`)

- Mensagens são JSON delimitado por `\n`, envelope `{id, type, payload}`.
- `HandlerManager.RegisterRequestHandler(type, fn)` recebe um handler tipado `func(ctx, Req) (Res, error)`; a resposta é enviada com o mesmo `id` e tipo `<type>_ok`.
- Erros: só `*tcp.ClientError` (criado com `tcp.NewClientError`) tem a mensagem repassada ao cliente; qualquer outro erro vira `"erro interno do servidor"` no tipo `error` e é logado.
- Cada conexão é um `Peer` com um `context` próprio carregando o peer; `tcp.SetUser(ctx, nome)` / `tcp.UserFrom(ctx)` associam o usuário autenticado à conexão. Escrita é assíncrona por uma `outbox` com buffer; se encher, a conexão é encerrada.
- `TCPServer` mantém o conjunto de peers conectados e oferece `Broadcast`; `TCPClient` (`Dial`) reutiliza o mesmo `Peer`/`HandlerManager` do lado cliente.

### Persistência

Schemas do ent em `ent/schema/` (User, UserPassword, Game, GameMove), todos com `BaseMixin` (id UUIDv7, `created_at`, `updated_at`). Use o helper `varchar` para strings com tamanho. O login registra o usuário automaticamente se o nome não existir (`infra/auth.go`). As partidas ativas ficam em memória no `infra.GameService`, além do registro no banco.

## Convenções

- Não escrever comentários no código Go.
- Mensagens de log e de erro em português; logs estruturados com `charm.land/log/v2` (`log.Info("mensagem", "chave", valor)`).
