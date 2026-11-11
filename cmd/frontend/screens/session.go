package screens

import (
	"xadrez/cmd/frontend/components"
	"xadrez/internal/application/messages"
	"xadrez/internal/infra/tcp"

	rg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const (
	buttonWidth  float32 = 200
	buttonHeight float32 = 40
	spacing      float32 = 10
)

type Session struct {
	size   rl.Vector2
	client *Client
	user   string
	board  *components.BoardRenderer
	invite *messages.InviteResponse
}

func (s *Session) Board() *components.BoardRenderer {
	if s.board == nil {
		s.board = components.NewBoardRenderer()
	}
	return s.board
}

func (s *Session) handleCommonEvent(event components.Event) (components.Screen, bool) {
	switch event.Type {
	case "room_invite":
		if invite, ok := event.Payload.(messages.InviteResponse); ok {
			s.invite = &invite
		}
		return nil, true
	case "join_room_ok", "create_room_ok":
		if room, ok := event.Payload.(messages.RoomResponse); ok {
			return NewRoomScreen(s, room), true
		}
		return nil, true
	}
	return nil, false
}

func (s *Session) acceptInvite() error {
	if s.invite == nil {
		return nil
	}
	room := s.invite.Room
	s.invite = nil
	return s.client.JoinRoom(messages.JoinRoomRequest{Room: room})
}

func (s *Session) renderInvite() (accepted bool) {
	if s.invite == nil {
		return false
	}
	bounds := rl.Rectangle{X: margin, Y: s.size.Y - margin - buttonHeight, Width: s.size.X - 2*margin, Height: buttonHeight}
	rl.DrawRectangleRec(bounds, rl.NewColor(230, 230, 250, 255))
	rl.DrawText(s.invite.From+" te convidou para a sala "+s.invite.Room, int32(bounds.X+spacing), int32(bounds.Y+11), textSize, rl.DarkBlue)

	buttonX := bounds.X + bounds.Width - 2*(100+spacing)
	if rg.Button(rl.Rectangle{X: buttonX, Y: bounds.Y + 5, Width: 100, Height: 30}, "Aceitar") {
		return true
	}
	if rg.Button(rl.Rectangle{X: buttonX + 100 + spacing, Y: bounds.Y + 5, Width: 100, Height: 30}, "Recusar") {
		s.invite = nil
	}
	return false
}

func errorMessage(event components.Event) (string, bool) {
	if event.Type != tcp.ErrorMessageType {
		return "", false
	}
	payload, ok := event.Payload.(tcp.ErrorPayload)
	if !ok {
		return "erro desconhecido", true
	}
	return payload.Message, true
}

func drawTitle(text string) {
	rl.DrawText(text, int32(margin), int32(margin), 28, rl.DarkGray)
}

func drawError(message string, y float32) {
	if message != "" {
		rl.DrawText(message, int32(margin), int32(y), textSize, rl.Red)
	}
}
