package main

import (
	"context"
	"xadrez/cmd/client/components"
	"xadrez/cmd/client/screens"
	"xadrez/internal/application/messages"
	"xadrez/internal/infra/tcp"

	"charm.land/log/v2"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	serverAddress = "localhost:9000"
	windowTitle   = "Xadrez"
	windowWidth   = 800
	windowHeight  = 600
)

func main() {
	events := make(chan components.Event, 64)

	hm := tcp.NewHandlerManager()
	forwardTCPMessageAsScreenEvent[messages.LoginResponse](hm, events, "login_ok")
	forwardTCPMessageAsScreenEvent[messages.ListOnlineUsersResponse](hm, events, "list_online_users_ok")
	forwardTCPMessageAsScreenEvent[messages.ListRoomsResponse](hm, events, "list_rooms_ok")
	forwardTCPMessageAsScreenEvent[messages.RoomResponse](hm, events, "create_room_ok")
	forwardTCPMessageAsScreenEvent[messages.RoomResponse](hm, events, "join_room_ok")
	forwardTCPMessageAsScreenEvent[messages.InviteResponse](hm, events, "invite_to_room_ok")
	forwardTCPMessageAsScreenEvent[messages.InviteResponse](hm, events, "room_invite")
	forwardTCPMessageAsScreenEvent[messages.MatchResponse](hm, events, "start_match_ok")
	forwardTCPMessageAsScreenEvent[messages.MatchResponse](hm, events, "get_match_ok")
	forwardTCPMessageAsScreenEvent[messages.MatchResponse](hm, events, "make_move_ok")
	forwardTCPMessageAsScreenEvent[messages.MatchResponse](hm, events, "resign_ok")
	forwardTCPMessageAsScreenEvent[messages.MatchResponse](hm, events, "match_updated")
	forwardTCPMessageAsScreenEvent[tcp.ErrorPayload](hm, events, "error")

	client := screens.NewClient(serverAddress, hm, events)
	defer client.Close()

	rl.InitWindow(windowWidth, windowHeight, windowTitle)
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	session := screens.NewSession(rl.Vector2{X: windowWidth, Y: windowHeight}, client)
	defer session.Close()

	var currentScreen components.Screen = screens.NewMainMenuScreen(session, "")

	for !rl.WindowShouldClose() && currentScreen != nil {
	drainEvents:
		for {
			select {
			case event := <-events:
				if event.Type == screens.DisconnectedEvent {
					log.Warn("conexão com o servidor perdida")
					currentScreen = screens.NewMainMenuScreen(session, "conexão com o servidor perdida")
					continue
				}
				currentScreen.HandleEvent(event)
			default:
				break drainEvents
			}
		}

		next, err := currentScreen.Update(context.Background())
		if err != nil {
			log.Error("erro ao atualizar a tela", "erro", err)
			return
		}

		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		currentScreen.Render()
		rl.EndDrawing()

		currentScreen = next
	}
}

func forwardTCPMessageAsScreenEvent[T any](hm *tcp.HandlerManager, events chan components.Event, messageType string) {
	hm.RegisterHandler(messageType, func(ctx context.Context, message tcp.Message[T]) error {
		events <- components.Event{
			Id:      message.Id,
			Type:    messageType,
			Payload: message.Payload,
		}
		return nil
	})
}
