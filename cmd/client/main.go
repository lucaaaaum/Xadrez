package main

import (
	"context"
	"flag"
	"xadrez/cmd/client/components"
	"xadrez/cmd/client/config"
	"xadrez/cmd/client/screens"
	"xadrez/internal/application/messages"
	"xadrez/internal/infra/tcp"

	"charm.land/log/v2"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	windowTitle     = "Xadrez"
	windowWidth     = 800
	windowHeight    = 600
	minWindowWidth  = 640
	minWindowHeight = 600
	eventBufferSize = 256
)

func main() {
	configPath := flag.String("config", config.DefaultPath(), "caminho do arquivo de configuração")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Warn("usando a configuração padrão", "erro", err)
	}

	events := make(chan components.Event, eventBufferSize)

	hm := tcp.NewHandlerManager()
	forwardTCPMessageAsScreenEvent[messages.LoginResponse](hm, events, "login_ok")
	forwardTCPMessageAsScreenEvent[messages.ListOnlineUsersResponse](hm, events, "list_online_users_ok")
	forwardTCPMessageAsScreenEvent[messages.ListRoomsResponse](hm, events, "list_rooms_ok")
	forwardTCPMessageAsScreenEvent[messages.RoomResponse](hm, events, "create_room_ok")
	forwardTCPMessageAsScreenEvent[messages.RoomResponse](hm, events, "join_room_ok")
	forwardTCPMessageAsScreenEvent[messages.InviteResponse](hm, events, "invite_to_room_ok")
	forwardTCPMessageAsScreenEvent[messages.InviteResponse](hm, events, "room_invite")
	forwardTCPMessageAsScreenEvent[messages.InviteResponse](hm, events, "decline_invite_ok")
	forwardTCPMessageAsScreenEvent[messages.InviteResponse](hm, events, "room_invite_declined")
	forwardTCPMessageAsScreenEvent[messages.MatchResponse](hm, events, "start_match_ok")
	forwardTCPMessageAsScreenEvent[messages.MatchResponse](hm, events, "get_match_ok")
	forwardTCPMessageAsScreenEvent[messages.MatchResponse](hm, events, "make_move_ok")
	forwardTCPMessageAsScreenEvent[messages.MatchResponse](hm, events, "resign_ok")
	forwardTCPMessageAsScreenEvent[messages.MatchResponse](hm, events, "match_updated")
	forwardTCPMessageAsScreenEvent[tcp.ErrorPayload](hm, events, "error")

	client := screens.NewClient(cfg.ServerAddress, hm, events)
	defer client.Close()

	rl.SetConfigFlags(rl.FlagWindowResizable)
	rl.InitWindow(windowWidth, windowHeight, windowTitle)
	defer rl.CloseWindow()

	rl.SetWindowMinSize(minWindowWidth, minWindowHeight)
	rl.SetTargetFPS(60)
	rl.SetExitKey(rl.KeyNull)

	session := screens.NewSession(client, cfg, *configPath)
	defer session.Close()

	var currentScreen components.Screen = screens.NewMainMenuScreen(session, "")
	switchTo := func(next components.Screen) {
		if next != currentScreen {
			client.NextGeneration()
		}
		currentScreen = next
	}

	for !rl.WindowShouldClose() && currentScreen != nil {
		session.Resize()

	drainEvents:
		for {
			select {
			case event := <-events:
				if event.Type == screens.DisconnectedEvent {
					log.Warn("conexão com o servidor perdida")
					switchTo(screens.NewMainMenuScreen(session, "conexão com o servidor perdida"))
					continue
				}
				if client.Settle(event) && event.Type == tcp.ErrorMessageType {
					log.Debug("erro de uma tela anterior descartado", "id", event.Id, "erro", event.Payload)
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

		switchTo(next)
	}
}

func forwardTCPMessageAsScreenEvent[T any](hm *tcp.HandlerManager, events chan components.Event, messageType string) {
	hm.RegisterHandler(messageType, func(ctx context.Context, message tcp.Message[T]) error {
		select {
		case events <- components.Event{Id: message.Id, Type: messageType, Payload: message.Payload}:
		default:
			log.Warn("fila de eventos cheia, mensagem descartada", "tipo", messageType, "id", message.Id)
		}
		return nil
	})
}
