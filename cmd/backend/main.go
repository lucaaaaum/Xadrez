package main

import (
	"context"
	"os"
	"xadrez/cmd/backend/handlers"
	"xadrez/ent"
	"xadrez/internal/domain/auth"
	"xadrez/internal/domain/chess"
	"xadrez/internal/domain/rooms"
	"xadrez/internal/domain/users"
	"xadrez/internal/infra"
	"xadrez/internal/infra/tcp"

	"charm.land/log/v2"
)

const defaultDatabaseDSN = "file:xadrez.db?_pragma=foreign_keys(1)"

type Services struct {
	loginService auth.LoginService
	userService  users.UserService
	roomService  rooms.RoomService
	gameService  chess.GameService
}

func newServices(client *ent.Client) Services {
	roomService := infra.NewRoomService(client)
	return Services{
		loginService: infra.NewAuthService(client),
		userService:  infra.NewUserService(client),
		roomService:  roomService,
		gameService:  infra.NewInMemoryGameService(),
	}
}

func registerHandlers(handlerManager *tcp.HandlerManager, services Services, notifier handlers.Notifier) {

	authHandler := handlers.NewAuthHandler(services.loginService, services.userService)
	handlerManager.RegisterRequestHandler("login", authHandler.HandleLogin)
	handlerManager.RegisterDisconnectHandler(authHandler.HandleDisconnect)

	userHandler := handlers.NewUserHandler(services.userService)
	handlerManager.RegisterRequestHandler("list_online_users", userHandler.HandleListOnlineUsers)

	roomHandler := handlers.NewRoomHandler(services.roomService, services.userService, notifier)
	handlerManager.RegisterRequestHandler("create_room", roomHandler.HandleCreateRoom)
	handlerManager.RegisterRequestHandler("join_room", roomHandler.HandleJoinRoom)
	handlerManager.RegisterRequestHandler("list_rooms", roomHandler.HandleListRooms)
	handlerManager.RegisterRequestHandler("invite_to_room", roomHandler.HandleInviteToRoom)

	chessHandler := handlers.NewChessHandler(services.gameService)
	handlerManager.RegisterRequestHandler("create_game", chessHandler.HandleCreateGameRequest)
	handlerManager.RegisterRequestHandler("list_games", chessHandler.HandleListGamesRequest)
	handlerManager.RegisterRequestHandler("get_board", chessHandler.HandleGetBoardRequest)
	handlerManager.RegisterRequestHandler("make_move", chessHandler.HandleMakeMoveRequest)

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
	handlerManager := tcp.NewHandlerManager()
	server := tcp.NewTCPServer(handlerManager)
	registerHandlers(handlerManager, services, server)
	if err := server.ListenAndServe(":9000"); err != nil {
		log.Fatal("erro no servidor", "erro", err)
	}
}
