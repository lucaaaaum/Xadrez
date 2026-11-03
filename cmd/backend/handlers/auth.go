package handlers

import (
	"context"
	"xadrez/internal/application/messages"
	"xadrez/internal/domain/auth"
	"xadrez/internal/infra/tcp"

	"charm.land/log/v2"
)

type AuthHandler struct {
	loginService auth.LoginService
}

func NewAuthHandler(loginService auth.LoginService) *AuthHandler {
	return &AuthHandler{
		loginService: loginService,
	}
}

func (h *AuthHandler) HandleLogin(ctx context.Context, request messages.LoginRequest) (messages.LoginResponse, error) {
	token, err := h.loginService.Login(ctx, request.Name, request.Password)
	if err != nil {
		log.Warn("falha no login", "usuário", request.Name, "erro", err)
		return messages.LoginResponse{}, tcp.NewClientError("usuário ou senha inválidos")
	}

	tcp.SetUser(ctx, request.Name)
	return messages.LoginResponse{Token: token}, nil
}
