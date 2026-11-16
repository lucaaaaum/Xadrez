package handlers

import (
	"context"
	"errors"
	"xadrez/internal/application/messages"
	"xadrez/internal/domain/auth"
	"xadrez/internal/domain/users"
	"xadrez/internal/infra/tcp"

	"charm.land/log/v2"
)

type AuthHandler struct {
	loginService auth.LoginService
	userService  users.UserService
}

func NewAuthHandler(loginService auth.LoginService, userService users.UserService) *AuthHandler {
	return &AuthHandler{
		loginService: loginService,
		userService:  userService,
	}
}

func (h *AuthHandler) HandleLogin(ctx context.Context, request messages.LoginRequest) (messages.LoginResponse, error) {
	if current, ok := tcp.UserFrom(ctx); ok {
		return messages.LoginResponse{}, tcp.NewClientError("a conexão já está autenticada como %s", current)
	}

	token, err := h.loginService.Login(ctx, request.Name, request.Password)
	if err != nil {
		log.Warn("falha no login", "usuário", request.Name, "erro", err)
		return messages.LoginResponse{}, tcp.NewClientError("usuário ou senha inválidos")
	}

	if err := h.userService.RegisterUserLogin(ctx, users.User{Name: request.Name}); err != nil {
		if errors.Is(err, users.ErrAlreadyOnline) {
			return messages.LoginResponse{}, tcp.NewClientError("o usuário %s já está logado", request.Name)
		}
		return messages.LoginResponse{}, err
	}

	tcp.SetUser(ctx, request.Name)
	return messages.LoginResponse{Token: token}, nil
}

func (h *AuthHandler) HandleDisconnect(ctx context.Context) {
	name, ok := tcp.UserFrom(ctx)
	if !ok {
		return
	}
	if err := h.userService.RegisterUserLogout(ctx, users.User{Name: name}); err != nil {
		log.Error("falha ao registrar logout", "usuário", name, "erro", err)
		return
	}
	log.Info("usuário desconectado", "usuário", name)
}
