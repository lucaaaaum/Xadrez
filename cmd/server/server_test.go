package main

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"slices"
	"strconv"
	"testing"
	"time"

	"xadrez/internal/application/messages"
	"xadrez/internal/domain/chess"
	"xadrez/internal/infra"
	"xadrez/internal/infra/tcp"
)

const readTimeout = 5 * time.Second

var foolsMate = [][4]int{
	{5, 1, 5, 2},
	{4, 6, 4, 4},
	{6, 1, 6, 3},
	{3, 7, 7, 3},
}

var fastestStalemate = [][4]int{
	{4, 1, 4, 2}, {0, 6, 0, 4},
	{3, 0, 7, 4}, {0, 7, 0, 5},
	{7, 4, 0, 4}, {7, 6, 7, 4},
	{7, 1, 7, 3}, {0, 5, 7, 5},
	{0, 4, 2, 6}, {5, 6, 5, 5},
	{2, 6, 3, 6}, {4, 7, 5, 6},
	{3, 6, 1, 6}, {3, 7, 3, 2},
	{1, 6, 1, 7}, {3, 2, 7, 6},
	{1, 7, 2, 7}, {5, 6, 6, 5},
	{2, 7, 4, 5},
}

func startServer(t *testing.T) string {
	t.Helper()
	client, err := infra.OpenDatabase(context.Background(), "file::memory:?_pragma=foreign_keys(1)")
	if err != nil {
		t.Fatalf("falha ao abrir o banco em memória: %v", err)
	}
	handlerManager := tcp.NewHandlerManager()
	server := tcp.NewTCPServer(handlerManager)
	registerHandlers(handlerManager, newServices(client), server)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	go server.Serve(ln)
	t.Cleanup(func() {
		ln.Close()
		client.Close()
	})
	return ln.Addr().String()
}

type testClient struct {
	t      *testing.T
	conn   net.Conn
	reader *bufio.Reader
	nextID int
	events []tcp.RawMessage
}

func connect(t *testing.T, address string) *testClient {
	t.Helper()
	conn, err := net.Dial("tcp", address)
	if err != nil {
		t.Fatalf("falha ao conectar: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	return &testClient{t: t, conn: conn, reader: bufio.NewReader(conn)}
}

func login(t *testing.T, address, name string) *testClient {
	t.Helper()
	c := connect(t, address)
	c.ok("login", messages.LoginRequest{Name: name, Password: "senha-" + name}, nil)
	return c
}

func (c *testClient) read() tcp.RawMessage {
	c.t.Helper()
	c.conn.SetReadDeadline(time.Now().Add(readTimeout))
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		c.t.Fatalf("falha ao ler do servidor: %v", err)
	}
	var message tcp.RawMessage
	if err := json.Unmarshal(line, &message); err != nil {
		c.t.Fatalf("mensagem inválida %q: %v", line, err)
	}
	return message
}

func (c *testClient) request(messageType string, payload any) tcp.RawMessage {
	c.t.Helper()
	c.nextID++
	id := strconv.Itoa(c.nextID)
	line, err := json.Marshal(tcp.Message[any]{Id: id, Type: messageType, Payload: payload})
	if err != nil {
		c.t.Fatal(err)
	}
	if _, err := c.conn.Write(append(line, '\n')); err != nil {
		c.t.Fatalf("falha ao enviar %s: %v", messageType, err)
	}
	for {
		message := c.read()
		if message.Id == id {
			return message
		}
		c.events = append(c.events, message)
	}
}

func (c *testClient) ok(messageType string, payload, response any) {
	c.t.Helper()
	message := c.request(messageType, payload)
	if message.Type != messageType+"_ok" {
		c.t.Fatalf("esperado %s_ok, obtido %s: %s", messageType, message.Type, message.Payload)
	}
	if response != nil {
		if err := json.Unmarshal(message.Payload, response); err != nil {
			c.t.Fatal(err)
		}
	}
}

func (c *testClient) fail(messageType string, payload any) string {
	c.t.Helper()
	message := c.request(messageType, payload)
	if message.Type != tcp.ErrorMessageType {
		c.t.Fatalf("esperado erro em %s, obtido %s: %s", messageType, message.Type, message.Payload)
	}
	var errorPayload tcp.ErrorPayload
	if err := json.Unmarshal(message.Payload, &errorPayload); err != nil {
		c.t.Fatal(err)
	}
	return errorPayload.Message
}

func (c *testClient) event(messageType string, payload any) {
	c.t.Helper()
	for i, message := range c.events {
		if message.Type == messageType {
			c.events = slices.Delete(c.events, i, i+1)
			c.decode(message, payload)
			return
		}
	}
	for {
		message := c.read()
		if message.Type == messageType {
			c.decode(message, payload)
			return
		}
		c.events = append(c.events, message)
	}
}

func (c *testClient) decode(message tcp.RawMessage, payload any) {
	c.t.Helper()
	if payload == nil {
		return
	}
	if err := json.Unmarshal(message.Payload, payload); err != nil {
		c.t.Fatal(err)
	}
}

func onlineUsers(c *testClient) []string {
	c.t.Helper()
	var response messages.ListOnlineUsersResponse
	c.ok("list_online_users", messages.ListOnlineUsersRequest{}, &response)
	return response.Users
}

func TestRequestsRequireLogin(t *testing.T) {
	address := startServer(t)
	c := connect(t, address)

	for messageType, payload := range map[string]any{
		"list_online_users": messages.ListOnlineUsersRequest{},
		"create_room":       messages.CreateRoomRequest{Room: "sala"},
		"make_move":         messages.MoveRequest{Room: "sala"},
	} {
		if text := c.fail(messageType, payload); text != "é necessário fazer login" {
			t.Fatalf("%s sem login: mensagem inesperada %q", messageType, text)
		}
	}
}

func TestLoginAndOnlineUsers(t *testing.T) {
	address := startServer(t)
	ana := login(t, address, "ana")
	login(t, address, "bob")

	if users := onlineUsers(ana); !slices.Equal(users, []string{"ana", "bob"}) {
		t.Fatalf("usuários online inesperados: %v", users)
	}

	if text := ana.fail("login", messages.LoginRequest{Name: "ana", Password: "senha-ana"}); text != "a conexão já está autenticada como ana" {
		t.Fatalf("mensagem inesperada: %q", text)
	}
	other := connect(t, address)
	if text := other.fail("login", messages.LoginRequest{Name: "ana", Password: "errada"}); text != "usuário ou senha inválidos" {
		t.Fatalf("mensagem inesperada: %q", text)
	}
	if text := other.fail("login", messages.LoginRequest{Name: "ana", Password: "senha-ana"}); text != "o usuário ana já está logado" {
		t.Fatalf("mensagem inesperada: %q", text)
	}
}

func TestDisconnectRemovesOnlineUser(t *testing.T) {
	address := startServer(t)
	ana := login(t, address, "ana")
	bob := login(t, address, "bob")
	bob.conn.Close()

	deadline := time.Now().Add(readTimeout)
	for slices.Contains(onlineUsers(ana), "bob") {
		if time.Now().After(deadline) {
			t.Fatal("bob continuou online após desconectar")
		}
		time.Sleep(10 * time.Millisecond)
	}
	login(t, address, "bob")
}

func TestInviteDeclineAndAccept(t *testing.T) {
	address := startServer(t)
	ana := login(t, address, "ana")
	bob := login(t, address, "bob")

	ana.ok("create_room", messages.CreateRoomRequest{Room: "sala"}, nil)
	if text := ana.fail("invite_to_room", messages.InviteToRoomRequest{Room: "sala", User: "dave"}); text != "o usuário dave não está online" {
		t.Fatalf("mensagem inesperada: %q", text)
	}
	if text := bob.fail("invite_to_room", messages.InviteToRoomRequest{Room: "sala", User: "ana"}); text != "apenas o administrador pode convidar para a sala sala" {
		t.Fatalf("mensagem inesperada: %q", text)
	}

	ana.ok("invite_to_room", messages.InviteToRoomRequest{Room: "sala", User: "bob"}, nil)
	var invite messages.InviteResponse
	bob.event("room_invite", &invite)
	if invite != (messages.InviteResponse{Room: "sala", From: "ana", To: "bob"}) {
		t.Fatalf("convite inesperado: %+v", invite)
	}

	bob.ok("decline_invite", messages.DeclineInviteRequest{Room: "sala"}, nil)
	var declined messages.InviteResponse
	ana.event("room_invite_declined", &declined)
	if declined.To != "bob" {
		t.Fatalf("aviso de recusa inesperado: %+v", declined)
	}

	ana.ok("invite_to_room", messages.InviteToRoomRequest{Room: "sala", User: "bob"}, nil)
	bob.event("room_invite", nil)
	var room messages.RoomResponse
	bob.ok("join_room", messages.JoinRoomRequest{Room: "sala"}, &room)
	if room.Admin != "ana" || !slices.Equal(room.Participants, []string{"ana", "bob"}) {
		t.Fatalf("sala inesperada: %+v", room)
	}
}

type matchPlayers struct {
	white, black, spectator *testClient
}

func startMatch(t *testing.T, address string) matchPlayers {
	t.Helper()
	players := matchPlayers{
		white:     login(t, address, "ana"),
		black:     login(t, address, "bob"),
		spectator: login(t, address, "carol"),
	}
	players.white.ok("create_room", messages.CreateRoomRequest{Room: "sala"}, nil)
	players.black.ok("join_room", messages.JoinRoomRequest{Room: "sala"}, nil)
	players.spectator.ok("join_room", messages.JoinRoomRequest{Room: "sala"}, nil)
	players.white.ok("start_match", messages.StartMatchRequest{Room: "sala", White: "ana", Black: "bob"}, nil)
	players.black.event("match_updated", nil)
	players.spectator.event("match_updated", nil)
	return players
}

func (p matchPlayers) play(t *testing.T, moves [][4]int) messages.MatchResponse {
	t.Helper()
	var response messages.MatchResponse
	for i, m := range moves {
		mover, opponent := p.white, p.black
		if i%2 == 1 {
			mover, opponent = p.black, p.white
		}
		mover.ok("make_move", messages.MoveRequest{Room: "sala", FromX: m[0], FromY: m[1], ToX: m[2], ToY: m[3]}, &response)
		var update messages.MatchResponse
		opponent.event("match_updated", &update)
		if !sameBoard(update.Pieces, response.Pieces) {
			t.Fatalf("lance %d: tabuleiro do oponente diferente do autor", i+1)
		}
		p.spectator.event("match_updated", nil)
	}
	return response
}

func sameBoard(a, b [8][8]*messages.Piece) bool {
	for x := range a {
		for y := range a[x] {
			if (a[x][y] == nil) != (b[x][y] == nil) || (a[x][y] != nil && *a[x][y] != *b[x][y]) {
				return false
			}
		}
	}
	return true
}

func requireOutcome(t *testing.T, who string, response messages.MatchResponse, outcome chess.Outcome, winner string) {
	t.Helper()
	if response.Outcome != outcome || response.Winner != winner {
		t.Fatalf("%s: esperado %s com vencedor %q, obtido %s com vencedor %q", who, outcome, winner, response.Outcome, response.Winner)
	}
}

func TestMoveRulesOverNetwork(t *testing.T) {
	address := startServer(t)
	p := startMatch(t, address)

	if text := p.black.fail("make_move", messages.MoveRequest{Room: "sala", FromX: 4, FromY: 6, ToX: 4, ToY: 4}); text != "não é a vez das pretas" {
		t.Fatalf("mensagem inesperada: %q", text)
	}
	if text := p.spectator.fail("make_move", messages.MoveRequest{Room: "sala", FromX: 4, FromY: 1, ToX: 4, ToY: 3}); text != "o usuário carol não é jogador da partida na sala sala" {
		t.Fatalf("mensagem inesperada: %q", text)
	}
	if text := p.white.fail("make_move", messages.MoveRequest{Room: "sala", FromX: 4, FromY: 1, ToX: 4, ToY: 5}); text == "" {
		t.Fatal("lance ilegal deveria ser recusado")
	}

	response := p.play(t, foolsMate[:1])
	if response.Turn != chess.Black || response.Pieces[5][2] == nil {
		t.Fatalf("estado inesperado após o lance: turno %v", response.Turn)
	}
}

func TestMatchOutcomes(t *testing.T) {
	t.Run("xeque-mate", func(t *testing.T) {
		p := startMatch(t, startServer(t))
		last := p.play(t, foolsMate)
		requireOutcome(t, "bob", last, chess.OutcomeCheckmate, "bob")
	})

	t.Run("afogamento", func(t *testing.T) {
		p := startMatch(t, startServer(t))
		last := p.play(t, fastestStalemate)
		requireOutcome(t, "ana", last, chess.OutcomeStalemate, "")
	})

	t.Run("desistência", func(t *testing.T) {
		p := startMatch(t, startServer(t))
		var response, update messages.MatchResponse
		p.black.ok("resign", messages.ResignRequest{Room: "sala"}, &response)
		p.white.event("match_updated", &update)
		requireOutcome(t, "bob", response, chess.OutcomeResignation, "ana")
		requireOutcome(t, "ana", update, chess.OutcomeResignation, "ana")
	})

	t.Run("desconexão", func(t *testing.T) {
		p := startMatch(t, startServer(t))
		p.black.conn.Close()
		var update messages.MatchResponse
		p.white.event("match_updated", &update)
		requireOutcome(t, "ana", update, chess.OutcomeDisconnection, "ana")
		p.spectator.event("match_updated", &update)
		requireOutcome(t, "carol", update, chess.OutcomeDisconnection, "ana")
	})
}

func TestNewMatchAfterFinish(t *testing.T) {
	address := startServer(t)
	p := startMatch(t, address)

	if text := p.white.fail("start_match", messages.StartMatchRequest{Room: "sala", White: "bob", Black: "carol"}); text != "já existe uma partida em andamento na sala sala" {
		t.Fatalf("mensagem inesperada: %q", text)
	}
	p.play(t, foolsMate)

	var response messages.MatchResponse
	p.white.ok("start_match", messages.StartMatchRequest{Room: "sala", White: "bob", Black: "carol"}, &response)
	if response.White != "bob" || response.Black != "carol" || response.Outcome != "" {
		t.Fatalf("nova partida inesperada: %+v", response)
	}
}
