package main

import (
	"context"
	"xadrez/cmd/frontend/components"
	"xadrez/cmd/frontend/screens"
	"xadrez/internal/application/messages"
	"xadrez/internal/infra/tcp"

	rl "github.com/gen2brain/raylib-go/raylib"
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

	tcpClient, err := tcp.Dial("localhost:9000", hm)
	if err != nil {
		panic(err)
	}

	rl.InitWindow(800, 600, "Hello World")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	client := screens.NewClient(tcpClient)
	var currentScreen components.Screen = screens.NewMainMenuScreen(rl.Vector2{X: 800, Y: 600}, client)

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		if currentScreen == nil {
			break
		}

	drainEvents:
		for {
			select {
			case event := <-events:
				currentScreen.HandleEvent(event)
			default:
				break drainEvents
			}
		}

		next, err := currentScreen.Update(context.Background())
		if err != nil {
			panic(err)
		}
		currentScreen.Render()
		currentScreen = next

		rl.EndDrawing()
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
