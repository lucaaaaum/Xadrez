package screens

import (
	"context"
	"xadrez/cmd/frontend/components"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type LobbyScreen struct {
	session      *Session
	buttons      *components.ButtonList
	next         components.Screen
	accept       bool
	errorMessage string
}

func NewLobbyScreen(session *Session) *LobbyScreen {
	l := &LobbyScreen{session: session}
	l.next = l

	l.buttons = components.NewButtonList(rl.Vector2{X: margin, Y: 80}, rl.Vector2{X: buttonWidth, Y: session.size.Y - 100})
	l.buttons.AddNewButton("Usuários online", func() {
		l.next = NewOnlineUsersScreen(session)
	})
	l.buttons.AddNewButton("Salas", func() {
		l.next = NewRoomsScreen(session)
	})
	l.buttons.AddNewButton("Sair", func() {
		l.next = nil
	})
	return l
}

func (l *LobbyScreen) HandleEvent(event components.Event) {
	if message, ok := errorMessage(event); ok {
		l.errorMessage = message
		return
	}
	if next, ok := l.session.handleCommonEvent(event); ok && next != nil {
		l.next = next
	}
}

func (l *LobbyScreen) Update(ctx context.Context) (components.Screen, error) {
	l.buttons.Update()
	if l.accept {
		l.accept = false
		if err := l.session.acceptInvite(); err != nil {
			l.errorMessage = err.Error()
		}
	}
	return l.next, nil
}

func (l *LobbyScreen) Render() error {
	drawTitle("Olá, " + l.session.user)
	l.buttons.Render()
	drawError(l.errorMessage, 80+3*60)
	if l.session.renderInvite() {
		l.accept = true
	}
	return nil
}
