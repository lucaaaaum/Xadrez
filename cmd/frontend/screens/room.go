package screens

import (
	"context"
	"strings"
	"xadrez/cmd/frontend/components"
	"xadrez/internal/application/messages"

	rg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type RoomScreen struct {
	session       *Session
	next          components.Screen
	room          messages.RoomResponse
	scroll        int32
	active        int32
	white         int32
	black         int32
	inviteName    string
	editingInvite bool
	infoMessage   string
	errorMessage  string

	refresh bool
	invite  bool
	start   bool
	watch   bool
	back    bool
	accept  bool
}

func NewRoomScreen(session *Session, room messages.RoomResponse) *RoomScreen {
	s := &RoomScreen{session: session, room: room, active: -1, black: 1}
	s.next = s
	return s
}

func (s *RoomScreen) isAdmin() bool {
	return s.room.Admin == s.session.user
}

func (s *RoomScreen) HandleEvent(event components.Event) {
	if message, ok := errorMessage(event); ok {
		s.errorMessage = message
		s.infoMessage = ""
		return
	}
	if next, ok := s.session.handleCommonEvent(event); ok {
		if next != nil {
			s.next = next
		}
		return
	}
	switch event.Type {
	case "list_rooms_ok":
		if response, ok := event.Payload.(messages.ListRoomsResponse); ok {
			for _, room := range response.Rooms {
				if room.Room == s.room.Room {
					s.room = room
				}
			}
		}
	case "invite_to_room_ok":
		if response, ok := event.Payload.(messages.InviteResponse); ok {
			s.infoMessage = "convite enviado para " + response.To
			s.errorMessage = ""
		}
	case "start_match_ok", "get_match_ok", "match_updated":
		if match, ok := event.Payload.(messages.MatchResponse); ok && match.Room == s.room.Room {
			s.next = s.matchScreen(match)
		}
	}
}

func (s *RoomScreen) matchScreen(match messages.MatchResponse) *MatchScreen {
	room := s.room
	return NewMatchScreen(s.session.size, s.session.user, match, s.session.client, s.session.Board(), func() components.Screen {
		return NewRoomScreen(s.session, room)
	})
}

func (s *RoomScreen) Update(ctx context.Context) (components.Screen, error) {
	var err error
	switch {
	case s.refresh:
		err = s.session.client.ListRooms()
	case s.invite:
		name := strings.TrimSpace(s.inviteName)
		if name == "" {
			s.errorMessage = "informe quem convidar"
			break
		}
		err = s.session.client.InviteToRoom(messages.InviteToRoomRequest{Room: s.room.Room, User: name})
	case s.start:
		white, black := s.participant(s.white), s.participant(s.black)
		if white == "" || black == "" || white == black {
			s.errorMessage = "escolha dois jogadores diferentes"
			break
		}
		err = s.session.client.StartMatch(messages.StartMatchRequest{Room: s.room.Room, White: white, Black: black})
	case s.watch:
		err = s.session.client.GetMatch(messages.GetMatchRequest{Room: s.room.Room})
	case s.accept:
		err = s.session.acceptInvite()
	case s.back:
		s.next = NewRoomsScreen(s.session)
	}
	s.refresh, s.invite, s.start, s.watch, s.accept, s.back = false, false, false, false, false, false
	if err != nil {
		s.errorMessage = err.Error()
	}
	return s.next, nil
}

func (s *RoomScreen) participant(index int32) string {
	if index < 0 || int(index) >= len(s.room.Participants) {
		return ""
	}
	return s.room.Participants[index]
}

func (s *RoomScreen) Render() error {
	drawTitle("Sala " + s.room.Room)
	rl.DrawText("Administrador: "+s.room.Admin, int32(margin), 55, textSize, rl.Gray)

	listHeight := s.session.size.Y - 230
	rg.ListView(rl.Rectangle{X: margin, Y: 85, Width: s.session.size.X - 3*margin - buttonWidth, Height: listHeight}, strings.Join(s.room.Participants, ";"), &s.scroll, &s.active)

	x := s.session.size.X - margin - buttonWidth
	y := float32(85)
	button := func(text string) bool {
		pressed := rg.Button(rl.Rectangle{X: x, Y: y, Width: buttonWidth, Height: buttonHeight}, text)
		y += buttonHeight + spacing
		return pressed
	}
	label := func(text string) {
		rl.DrawText(text, int32(x), int32(y), textSize, rl.DarkGray)
		y += float32(lineHeight)
	}

	if button("Atualizar") {
		s.refresh = true
	}
	if button("Assistir partida") {
		s.watch = true
	}

	if s.isAdmin() {
		label("Convidar usuário")
		if rg.TextBox(rl.Rectangle{X: x, Y: y, Width: buttonWidth, Height: buttonHeight}, &s.inviteName, 32, s.editingInvite) {
			s.editingInvite = !s.editingInvite
		}
		y += buttonHeight + spacing
		if button("Convidar") {
			s.invite = true
		}

		if len(s.room.Participants) > 0 {
			options := strings.Join(s.room.Participants, ";")
			label("Brancas")
			rg.ComboBox(rl.Rectangle{X: x, Y: y, Width: buttonWidth, Height: 30}, options, &s.white)
			y += 30 + spacing
			label("Pretas")
			rg.ComboBox(rl.Rectangle{X: x, Y: y, Width: buttonWidth, Height: 30}, options, &s.black)
			y += 30 + spacing
			if button("Iniciar partida") {
				s.start = true
			}
		}
	}
	if button("Voltar") {
		s.back = true
	}

	messageY := 85 + listHeight + spacing
	if s.infoMessage != "" {
		rl.DrawText(s.infoMessage, int32(margin), int32(messageY), textSize, rl.DarkGreen)
	}
	drawError(s.errorMessage, messageY)
	if s.session.renderInvite() {
		s.accept = true
	}
	return nil
}
