package main

import (
	"context"
	"os"
	"xadrez/cmd/backend/handlers"
	"xadrez/ent"
	"xadrez/internal/domain/auth"
	"xadrez/internal/domain/chess"
	"xadrez/internal/domain/users"
	"xadrez/internal/infra"
	"xadrez/internal/infra/tcp"

	"charm.land/log/v2"
)

const defaultDatabaseDSN = "file:xadrez.db?_pragma=foreign_keys(1)"

type Services struct {
	loginService auth.LoginService
	userService  users.UserService
	gameService  chess.GameService
}

func newServices(client *ent.Client) Services {
	return Services{
		loginService: infra.NewAuthService(client),
		userService:  infra.NewInMemoryUserService(),
		gameService:  infra.NewInMemoryGameService(),
	}
}

func newHandlerManager(services Services) *tcp.HandlerManager {
	handlerManager := tcp.NewHandlerManager()

	authHandler := handlers.NewAuthHandler(services.loginService)
	handlerManager.RegisterRequestHandler("login", authHandler.HandleLogin)

	chessHandler := handlers.NewChessHandler(services.gameService)
	handlerManager.RegisterRequestHandler("create_game", chessHandler.HandleCreateGameRequest)
	handlerManager.RegisterRequestHandler("list_games", chessHandler.HandleListGamesRequest)
	handlerManager.RegisterRequestHandler("get_board", chessHandler.HandleGetBoardRequest)
	handlerManager.RegisterRequestHandler("make_move", chessHandler.HandleMakeMoveRequest)

	return handlerManager
}

func main() {
	dsn := os.Getenv("XADREZ_DATABASE_DSN")
	if dsn == "" {
		dsn = defaultDatabaseDSN
	}
	client, err := infra.OpenDatabase(context.Background(), dsn)
	if err != nil {
		log.Fatal("erro ao abrir o banco de dados", "erro", err)
	}
	defer client.Close()

	services := newServices(client)
	server := tcp.NewTCPServer(newHandlerManager(services))
	if err := server.ListenAndServe(":9000"); err != nil {
		log.Fatal("erro no servidor", "erro", err)
	}
}
