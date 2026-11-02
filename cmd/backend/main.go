package main

import (
	"xadrez/cmd/backend/handlers"
	"xadrez/internal/domain/auth"
	"xadrez/internal/domain/users"
	"xadrez/internal/infra"
	"xadrez/internal/infra/tcp"

	"charm.land/log/v2"
)

type Services struct {
	loginService auth.LoginService
	userService  users.UserService
}

func newServices() Services {
	return Services{
		loginService: infra.NewInMemoryAuthService(),
		userService:  infra.NewInMemoryUserService(),
	}
}

func newHandlerManager(services Services) *tcp.HandlerManager {
	handlerManager := tcp.NewHandlerManager()

	authHandler := handlers.NewAuthHandler(services.loginService)
	handlerManager.RegisterRequestHandler("login", authHandler.HandleLogin)

	return handlerManager
}

func main() {
	services := newServices()
	server := tcp.NewTCPServer(newHandlerManager(services))
	if err := server.ListenAndServe(":9000"); err != nil {
		log.Fatal("erro no servidor", "erro", err)
	}
}
