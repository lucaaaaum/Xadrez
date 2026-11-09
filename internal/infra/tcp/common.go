package tcp

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"charm.land/log/v2"
)

const ErrorMessageType = "error"

type RawMessage struct {
	Id      string          `json:"id,omitempty"`
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload,omitempty"`
}

type Message[T any] struct {
	Id      string `json:"id,omitempty"`
	Type    string `json:"type"`
	Payload T      `json:"payload"`
}

type ErrorPayload struct {
	Message string `json:"message"`
}

type ClientError struct {
	Message string
}

func (e *ClientError) Error() string {
	return e.Message
}

func NewClientError(format string, args ...any) *ClientError {
	return &ClientError{Message: fmt.Sprintf(format, args...)}
}

type HandlerManager struct {
	handlers           map[string]func(ctx context.Context, message RawMessage) error
	disconnectHandlers []func(ctx context.Context)
}

func NewHandlerManager() *HandlerManager {
	return &HandlerManager{
		handlers: make(map[string]func(ctx context.Context, message RawMessage) error),
	}
}

func (m *HandlerManager) RegisterHandler[T any](messageType string, handleFunc func(ctx context.Context, message Message[T]) error) {
	log.Info("registrando handler", "tipo", messageType)
	m.handlers[messageType] = func(ctx context.Context, message RawMessage) error {
		var typedMessage Message[T]
		typedMessage.Id = message.Id
		typedMessage.Type = message.Type
		if err := json.Unmarshal(message.Payload, &typedMessage.Payload); err != nil {
			log.Warn("falha ao desserializar payload", "tipo", message.Type, "erro", err)
			return NewClientError("payload inválido para o tipo de mensagem %q", message.Type)
		}
		return handleFunc(ctx, typedMessage)
	}
}

func (m *HandlerManager) RegisterRequestHandler[Req, Res any](messageType string, handleFunc func(ctx context.Context, request Req) (Res, error)) {
	m.RegisterHandler(messageType, func(ctx context.Context, message Message[Req]) error {
		response, err := handleFunc(ctx, message.Payload)
		if err != nil {
			log.Warn("falha ao processar requisição", "tipo", messageType, "erro", err)
			return err
		}
		return peerFrom(ctx).Reply(message.Id, messageType+"_ok", response)
	})
}

func (m *HandlerManager) RegisterDisconnectHandler(handleFunc func(ctx context.Context)) {
	m.disconnectHandlers = append(m.disconnectHandlers, handleFunc)
}

func (m *HandlerManager) HandleDisconnect(peer *Peer) {
	for _, handleFunc := range m.disconnectHandlers {
		handleFunc(peer.ctx)
	}
}

func (m *HandlerManager) HandleMessage(peer *Peer, jsonPayload []byte) {
	var rawMessage RawMessage
	if err := json.Unmarshal(jsonPayload, &rawMessage); err != nil {
		log.Warn("falha ao desserializar mensagem", "cliente", peer.RemoteAddr(), "erro", err)
		peer.replyError("", NewClientError("mensagem malformada"))
		return
	}

	handler, ok := m.handlers[rawMessage.Type]
	if !ok {
		log.Warn("tipo de mensagem desconhecido", "cliente", peer.RemoteAddr(), "tipo", rawMessage.Type)
		peer.replyError(rawMessage.Id, NewClientError("tipo de mensagem desconhecido: %q", rawMessage.Type))
		return
	}

	if err := handler(peer.ctx, rawMessage); err != nil {
		peer.replyError(rawMessage.Id, err)
	}
}

func (p *Peer) replyError(id string, err error) {
	var clientError *ClientError
	if !errors.As(err, &clientError) {
		log.Error("erro ao processar mensagem", "cliente", p.RemoteAddr(), "id", id, "erro", err)
		clientError = NewClientError("erro interno do servidor")
	}
	if err := p.Reply(id, ErrorMessageType, ErrorPayload{Message: clientError.Message}); err != nil {
		log.Error("erro ao enviar mensagem de erro", "cliente", p.RemoteAddr(), "erro", err)
	}
}
