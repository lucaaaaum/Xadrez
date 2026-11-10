# Roadmap

## Servidor

### Conexão

- [x] Servidor TCP que aceita várias conexões simultâneas.
- [x] Porta lida da configuração.
- [x] Sessão por conexão, associando a conexão ao usuário autenticado.
- [x] Detecção de desconexão, com o usuário removido da lista de online.
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
- [x] Validação de que só o jogador da vez move as próprias peças.
- [ ] Limite de tentativas de login.

### Correções

- [x] Import inexistente de `uuid` em `internal/infra/auth.go`.
- [x] Receivers por ponteiro nos serviços em memória, para que o estado persista.
- [x] Acesso concorrente seguro ao estado compartilhado.
- [x] Remoção do usuário em `RegisterUserLogout`.

## Usuários

### Cadastro

- [ ] Cadastro com nome de usuário e senha, separado do login.
- [x] Nome de usuário único.

### Autenticação

- [x] Login.
- [x] Logout.

### Online

- [x] Listagem de usuários online, visível a todos os usuários logados.
- [ ] Notificação aos clientes quando alguém entra ou sai.

## Partidas

### Salas

- [x] Sala (`GameRoom`) com administrador e participantes, separada das partidas (`Match`).
- [x] Partida com referência aos participantes branco e preto, e lances ligados à partida.
- [x] Criador da sala definido como administrador.
- [x] Entrar numa sala como espectador.
- [ ] Sair da sala.
- [x] Administrador escolhe os jogadores branco e preto e inicia a partida.
- [x] Uma partida por vez na sala.
- [x] Nova partida na mesma sala após o fim da anterior.
- [ ] Administrador bane um participante.
- [ ] Banido impedido de entrar de novo.
- [ ] Transferência de administrador quando ele sai.

### Convites

- [x] Administrador convida qualquer usuário online para a sala (sem lista de amigos).
- [x] Convite entregue ao convidado em tempo real.
- [ ] Aceitar convite, entrando na sala como participante.
- [ ] Recusar convite, com aviso ao administrador.
- [ ] Cancelar convite pendente.
- [ ] Convite expirado quando a sala fecha ou o convidado desconecta.
- [ ] Listagem dos convites pendentes do usuário.

### Jogo

- [x] Decidir: biblioteca de xadrez ou implementação própria.
- [x] Movimento básico de cada peça.
- [x] Lances restritos aos limites do tabuleiro.
- [x] Caminho livre para bispo, torre, dama e avanço do peão.
- [x] Captura do peão na diagonal, sem captura no avanço reto.
- [x] Lance que deixa o próprio rei em xeque é rejeitado.
- [ ] Roque, sem passar por casa atacada e com rei e torre sem terem se movido.
- [ ] En passant.
- [x] Promoção, com a peça escolhida no `MoveRequest`.
- [ ] Turnos, com a fila de ações por partida.
- [x] Tabuleiro atualizado para jogadores e espectadores a cada lance.
- [x] Erros do domínio enviados ao cliente como erro de cliente.
- [x] Nome de sala único na criação.
- [ ] `HandleValidateMoveRequest` validando sem executar o lance.

### Fim de jogo

- [x] Xeque-mate.
- [x] Afogamento.
- [ ] Empate por 50 lances, repetição tripla e material insuficiente.
- [x] Resultado (vitória, derrota, empate) enviado aos dois jogadores.
- [x] Desistência, com vitória concedida ao oponente.
- [ ] Vitória por desconexão do oponente.

## Eventos

- [ ] Barramento de eventos no servidor.
- [ ] Log de eventos: cadastro, login, início e fim de partida, erros.

## Persistência

- [x] Banco de dados para usuários.
- [ ] Eventos do servidor gravados no banco.
- [x] Histórico de partidas.
- [x] Decidir: banco (SQLite por ora, com migração para Postgres se necessário).

## Cliente

- [ ] Cliente com raylib.
- [ ] Telas de cadastro e login.
- [ ] Lista de usuários online e convites.
- [ ] Tabuleiro jogável.
- [ ] Tela de resultado.

## Opcionais

- [ ] Ranking público por vitórias e empates.
- [ ] Jogadores sorteados de uma fila da sala.
- [ ] Partidas paralelas na mesma sala.
- [ ] Visibilidade da sala (pública ou privada).
- [ ] Senha para entrar na sala.
- [ ] Bot gerenciado pelo servidor.
- [ ] Reconexão a uma partida em andamento.
- [ ] Painel de administração em tempo real.
- [ ] Deploy na nuvem.

## Testes

- [x] Testes unitários das regras do xadrez.
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
