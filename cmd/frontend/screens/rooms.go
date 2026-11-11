package screens

import (
	"context"
	"fmt"
	"strings"
	"xadrez/cmd/frontend/components"
	"xadrez/internal/application/messages"

	rg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type RoomsScreen struct {
	session      *Session
	next         components.Screen
	rooms        []messages.RoomResponse
	scroll       int32
	active       int32
	newRoom      string
	editingRoom  bool
	errorMessage string

	refresh bool
	create  bool
	join    bool
	back    bool
	accept  bool
}

func NewRoomsScreen(session *Session) *RoomsScreen {
	s := &RoomsScreen{session: session, active: -1, refresh: true}
	s.next = s
	return s
}

func (s *RoomsScreen) HandleEvent(event components.Event) {
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
	if event.Type == "list_rooms_ok" {
		if response, ok := event.Payload.(messages.ListRoomsResponse); ok {
			s.rooms = response.Rooms
			s.errorMessage = ""
			if int(s.active) >= len(s.rooms) {
				s.active = -1
			}
		}
	}
}

func (s *RoomsScreen) Update(ctx context.Context) (components.Screen, error) {
	var err error
	switch {
	case s.refresh:
		err = s.session.client.ListRooms()
	case s.create:
		name := strings.TrimSpace(s.newRoom)
		if name == "" {
			s.errorMessage = "informe o nome da sala"
			break
		}
		err = s.session.client.CreateRoom(messages.CreateRoomRequest{Room: name})
	case s.join:
		if s.active < 0 || int(s.active) >= len(s.rooms) {
			s.errorMessage = "selecione uma sala"
			break
		}
		err = s.session.client.JoinRoom(messages.JoinRoomRequest{Room: s.rooms[s.active].Room})
	case s.accept:
		err = s.session.acceptInvite()
	case s.back:
		s.next = NewLobbyScreen(s.session)
	}
	s.refresh, s.create, s.join, s.accept, s.back = false, false, false, false, false
	if err != nil {
		s.errorMessage = err.Error()
	}
	return s.next, nil
}

func (s *RoomsScreen) Render() error {
	drawTitle("Salas")

	labels := make([]string, 0, len(s.rooms))
	for _, room := range s.rooms {
		labels = append(labels, fmt.Sprintf("%s (admin: %s, %d participantes)", room.Room, room.Admin, len(room.Participants)))
	}
	listHeight := s.session.size.Y - 200
	rg.ListView(rl.Rectangle{X: margin, Y: 70, Width: s.session.size.X - 3*margin - buttonWidth, Height: listHeight}, strings.Join(labels, ";"), &s.scroll, &s.active)

	x := s.session.size.X - margin - buttonWidth
	y := float32(70)
	button := func(text string) bool {
		pressed := rg.Button(rl.Rectangle{X: x, Y: y, Width: buttonWidth, Height: buttonHeight}, text)
		y += buttonHeight + spacing
		return pressed
	}

	if button("Entrar na sala") {
		s.join = true
	}
	if button("Atualizar") {
		s.refresh = true
	}
	y += spacing
	rl.DrawText("Nova sala", int32(x), int32(y), textSize, rl.DarkGray)
	y += float32(lineHeight)
	if rg.TextBox(rl.Rectangle{X: x, Y: y, Width: buttonWidth, Height: buttonHeight}, &s.newRoom, 32, s.editingRoom) {
		s.editingRoom = !s.editingRoom
	}
	y += buttonHeight + spacing
	if button("Criar sala") {
		s.create = true
	}
	y += spacing
	if button("Voltar") {
		s.back = true
	}

	drawError(s.errorMessage, 70+listHeight+spacing)
	if s.session.renderInvite() {
		s.accept = true
	}
	return nil
}
