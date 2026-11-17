package tcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"testing"
	"time"

	"xadrez/internal/domain"
)

type echoRequest struct {
	Text string `json:"text"`
}

type echoResponse struct {
	Text string `json:"text"`
	User string `json:"user"`
}

type pipeClient struct {
	t      *testing.T
	conn   net.Conn
	reader *bufio.Reader
}

func newPipe(t *testing.T, manager *HandlerManager) (*pipeClient, *Peer) {
	t.Helper()
	serverSide, clientSide := net.Pipe()
	peer := NewPeer(serverSide, manager)
	go peer.Run()
	t.Cleanup(func() {
		clientSide.Close()
		peer.Close()
	})
	return &pipeClient{t: t, conn: clientSide, reader: bufio.NewReader(clientSide)}, peer
}

func (c *pipeClient) sendRaw(line string) {
	c.t.Helper()
	c.conn.SetWriteDeadline(time.Now().Add(2 * time.Second))
	if _, err := c.conn.Write([]byte(line + "\n")); err != nil {
		c.t.Fatalf("falha ao enviar: %v", err)
	}
}

func (c *pipeClient) read() RawMessage {
	c.t.Helper()
	c.conn.SetReadDeadline(time.Now().Add(2 * time.Second))
	line, err := c.reader.ReadBytes('\n')
	if err != nil {
		c.t.Fatalf("falha ao ler resposta: %v", err)
	}
	var message RawMessage
	if err := json.Unmarshal(line, &message); err != nil {
		c.t.Fatalf("resposta inválida %q: %v", line, err)
	}
	return message
}

func (c *pipeClient) readError(id string) string {
	c.t.Helper()
	message := c.read()
	if message.Type != ErrorMessageType || message.Id != id {
		c.t.Fatalf("esperado erro com id %q, obtido %+v", id, message)
	}
	var payload ErrorPayload
	if err := json.Unmarshal(message.Payload, &payload); err != nil {
		c.t.Fatal(err)
	}
	return payload.Message
}

func newEchoManager() *HandlerManager {
	manager := NewHandlerManager()
	manager.RegisterRequestHandler("echo", func(ctx context.Context, request echoRequest) (echoResponse, error) {
		user, _ := UserFrom(ctx)
		return echoResponse{Text: request.Text, User: user}, nil
	})
	manager.RegisterRequestHandler("login", func(ctx context.Context, request echoRequest) (echoResponse, error) {
		SetUser(ctx, request.Text)
		return echoResponse{User: request.Text}, nil
	})
	manager.RegisterRequestHandler("client_error", func(ctx context.Context, request echoRequest) (echoResponse, error) {
		return echoResponse{}, NewClientError("erro do cliente %s", request.Text)
	})
	manager.RegisterRequestHandler("domain_error", func(ctx context.Context, request echoRequest) (echoResponse, error) {
		return echoResponse{}, domain.NewError("erro de domínio")
	})
	manager.RegisterRequestHandler("internal_error", func(ctx context.Context, request echoRequest) (echoResponse, error) {
		return echoResponse{}, errors.New("detalhe interno sigiloso")
	})
	return manager
}

func TestRequestGetsResponseWithSameId(t *testing.T) {
	client, _ := newPipe(t, newEchoManager())
	client.sendRaw(`{"id":"42","type":"echo","payload":{"text":"oi"}}`)

	message := client.read()
	if message.Id != "42" || message.Type != "echo_ok" {
		t.Fatalf("esperado echo_ok com id 42, obtido %+v", message)
	}
	var response echoResponse
	if err := json.Unmarshal(message.Payload, &response); err != nil || response.Text != "oi" {
		t.Fatalf("payload inesperado %s: %v", message.Payload, err)
	}
}

func TestInvalidMessagesGetErrors(t *testing.T) {
	client, _ := newPipe(t, newEchoManager())

	client.sendRaw(`isto não é json`)
	if text := client.readError(""); text != "mensagem malformada" {
		t.Fatalf("mensagem inesperada: %q", text)
	}

	client.sendRaw(`{"id":"1","type":"inexistente","payload":{}}`)
	if text := client.readError("1"); text != `tipo de mensagem desconhecido: "inexistente"` {
		t.Fatalf("mensagem inesperada: %q", text)
	}

	client.sendRaw(`{"id":"2","type":"echo","payload":{"text":10}}`)
	if text := client.readError("2"); text != `payload inválido para o tipo de mensagem "echo"` {
		t.Fatalf("mensagem inesperada: %q", text)
	}
}

func TestOnlyClientAndDomainErrorsReachClient(t *testing.T) {
	client, _ := newPipe(t, newEchoManager())

	client.sendRaw(`{"id":"1","type":"client_error","payload":{"text":"x"}}`)
	if text := client.readError("1"); text != "erro do cliente x" {
		t.Fatalf("mensagem inesperada: %q", text)
	}

	client.sendRaw(`{"id":"2","type":"domain_error","payload":{}}`)
	if text := client.readError("2"); text != "erro de domínio" {
		t.Fatalf("mensagem inesperada: %q", text)
	}

	client.sendRaw(`{"id":"3","type":"internal_error","payload":{}}`)
	if text := client.readError("3"); text != "erro interno do servidor" {
		t.Fatalf("detalhe interno vazou para o cliente: %q", text)
	}
}

func TestUserIsBoundToConnection(t *testing.T) {
	manager := newEchoManager()
	first, _ := newPipe(t, manager)
	second, _ := newPipe(t, manager)

	first.sendRaw(`{"id":"1","type":"login","payload":{"text":"ana"}}`)
	first.read()

	first.sendRaw(`{"id":"2","type":"echo","payload":{}}`)
	second.sendRaw(`{"id":"3","type":"echo","payload":{}}`)

	var response echoResponse
	json.Unmarshal(first.read().Payload, &response)
	if response.User != "ana" {
		t.Fatalf("a primeira conexão deveria estar autenticada como ana, está como %q", response.User)
	}
	json.Unmarshal(second.read().Payload, &response)
	if response.User != "" {
		t.Fatalf("a segunda conexão não deveria ter usuário, tem %q", response.User)
	}
}

func TestDisconnectHandlersReceiveUser(t *testing.T) {
	manager := newEchoManager()
	disconnected := make(chan string, 1)
	manager.RegisterDisconnectHandler(func(ctx context.Context) {
		user, _ := UserFrom(ctx)
		disconnected <- user
	})
	client, _ := newPipe(t, manager)

	client.sendRaw(`{"id":"1","type":"login","payload":{"text":"ana"}}`)
	client.read()
	client.conn.Close()

	select {
	case user := <-disconnected:
		if user != "ana" {
			t.Fatalf("handler de desconexão recebeu %q", user)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("handler de desconexão não foi chamado")
	}
}

func TestUserFromWithoutPeer(t *testing.T) {
	SetUser(context.Background(), "ana")
	if user, ok := UserFrom(context.Background()); ok || user != "" {
		t.Fatalf("contexto sem conexão não deveria ter usuário, obtido %q", user)
	}
}

func TestClosedPeerRejectsMessages(t *testing.T) {
	_, peer := newPipe(t, newEchoManager())
	peer.Close()

	if err := peer.SendMessage("echo", echoResponse{}); !errors.Is(err, ErrPeerClosed) {
		t.Fatalf("esperado ErrPeerClosed, obtido %v", err)
	}
	select {
	case <-peer.Done():
	default:
		t.Fatal("Done deveria estar fechado")
	}
}
