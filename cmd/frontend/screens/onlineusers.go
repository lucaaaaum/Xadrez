package screens

import (
	"context"
	"strings"
	"xadrez/cmd/frontend/components"
	"xadrez/internal/application/messages"

	rg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type OnlineUsersScreen struct {
	session      *Session
	next         components.Screen
	users        []string
	scroll       int32
	active       int32
	errorMessage string

	refresh bool
	back    bool
	accept  bool
}

func NewOnlineUsersScreen(session *Session) *OnlineUsersScreen {
	s := &OnlineUsersScreen{session: session, active: -1, refresh: true}
	s.next = s
	return s
}

func (s *OnlineUsersScreen) HandleEvent(event components.Event) {
	if message, ok := errorMessage(event); ok {
		s.errorMessage = message
		return
	}
	if next, ok := s.session.handleCommonEvent(event); ok {
		if next != nil {
			s.next = next
		}
		return
	}
	if event.Type == "list_online_users_ok" {
		if response, ok := event.Payload.(messages.ListOnlineUsersResponse); ok {
			s.users = response.Users
			s.errorMessage = ""
		}
	}
}

func (s *OnlineUsersScreen) Update(ctx context.Context) (components.Screen, error) {
	if s.refresh {
		s.refresh = false
		if err := s.session.client.ListOnlineUsers(); err != nil {
			s.errorMessage = err.Error()
		}
	}
	if s.accept {
		s.accept = false
		if err := s.session.acceptInvite(); err != nil {
			s.errorMessage = err.Error()
		}
	}
	if s.back {
		s.next = NewLobbyScreen(s.session)
	}
	return s.next, nil
}

func (s *OnlineUsersScreen) Render() error {
	drawTitle("Usuários online")

	listHeight := s.session.size.Y - 200
	rg.ListView(rl.Rectangle{X: margin, Y: 70, Width: s.session.size.X - 3*margin - buttonWidth, Height: listHeight}, strings.Join(s.users, ";"), &s.scroll, &s.active)

	x := s.session.size.X - margin - buttonWidth
	if rg.Button(rl.Rectangle{X: x, Y: 70, Width: buttonWidth, Height: buttonHeight}, "Atualizar") {
		s.refresh = true
	}
	if rg.Button(rl.Rectangle{X: x, Y: 70 + buttonHeight + spacing, Width: buttonWidth, Height: buttonHeight}, "Voltar") {
		s.back = true
	}

	drawError(s.errorMessage, 70+listHeight+spacing)
	if s.session.renderInvite() {
		s.accept = true
	}
	return nil
}
