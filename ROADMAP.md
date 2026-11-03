# Roadmap

## Servidor

### Conexão

- [x] Servidor TCP que aceita várias conexões simultâneas.
- [ ] Porta lida da configuração.
- [x] Sessão por conexão, associando a conexão ao usuário autenticado.
- [ ] Detecção de desconexão, com o usuário removido da lista de online.
- [ ] Graceful shutdown.

### Protocolo

- [x] Envelope de mensagem (JSON delimitado por linha, com `type` e `payload`).
- [ ] Tipos de mensagem de cliente e servidor.
- [x] Respostas de erro padronizadas.
- [ ] Leitura da resposta de uma requisição, correlacionada pelo `id`.
- [ ] Limite de tamanho de mensagem.

### Segurança

- [ ] TLS na conexão cliente-servidor.
- [x] Senhas com hash (Argon2id).
- [ ] Token de sessão aleatório (`crypto/rand`).
- [ ] Validação de que só o jogador da vez move as próprias peças.
- [ ] Limite de tentativas de login.

### Correções

- [x] Import inexistente de `uuid` em `internal/infra/auth.go`.
- [ ] Receivers por ponteiro nos serviços em memória, para que o estado persista.
- [ ] Acesso concorrente seguro ao estado compartilhado.
- [ ] Remoção do usuário em `RegisterUserLogout`.

## Usuários

### Cadastro

- [ ] Cadastro com nome de usuário e senha, separado do login.
- [ ] Nome de usuário único.

### Autenticação

- [ ] Login.
- [ ] Logout.

### Online

- [ ] Listagem de usuários online.
- [ ] Notificação aos clientes quando alguém entra ou sai.

## Partidas

### Convites

- [ ] Criar partida e convidar um usuário online.
- [ ] Aceitar ou recusar convite.
- [ ] Cancelar convite pendente.

### Jogo

- [ ] Regras do xadrez: validação de lances, roque, en passant, promoção.
- [ ] Turnos, com a fila de ações por partida.
- [ ] Tabuleiro atualizado para os dois jogadores a cada lance.
- [ ] Decidir: biblioteca de xadrez ou implementação própria.

### Fim de jogo

- [ ] Xeque-mate, afogamento e demais empates.
- [ ] Resultado (vitória, derrota, empate) enviado aos dois jogadores.
- [ ] Desistência, com vitória concedida ao oponente.
- [ ] Vitória por desconexão do oponente.

## Eventos

- [ ] Barramento de eventos no servidor.
- [ ] Log de eventos: cadastro, login, início e fim de partida, erros.

## Persistência

- [ ] Banco de dados para usuários.
- [ ] Eventos do servidor gravados no banco.
- [ ] Histórico de partidas.
- [ ] Decidir: banco (SQLite ou outro).

## Cliente

- [ ] Cliente com raylib.
- [ ] Telas de cadastro e login.
- [ ] Lista de usuários online e convites.
- [ ] Tabuleiro jogável.
- [ ] Tela de resultado.

## Opcionais

- [ ] Ranking público por vitórias e empates.
- [ ] Bot gerenciado pelo servidor.
- [ ] Reconexão a uma partida em andamento.
- [ ] Painel de administração em tempo real.
- [ ] Deploy na nuvem.

## Testes

- [ ] Testes unitários das regras do xadrez.
- [ ] Testes funcionais cobrindo todos os resultados de partida.
- [ ] Verificação com Wireshark de que o tráfego está cifrado e íntegro.
- [ ] Testes de robustez (man in the middle, DoS).

## Documentação

- [ ] Componentes de rede: cliente, servidor, banco, protocolo.
- [ ] Diagramas de sequência e de estado da comunicação.
- [ ] IPC e payloads das mensagens.
- [ ] Justificativa das bibliotecas de rede.
- [ ] Monitoramento e depuração da comunicação.
- [ ] Apresentação final.
