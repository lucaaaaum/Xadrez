package screens

import (
	"xadrez/cmd/client/components"
	"xadrez/cmd/client/config"
	"xadrez/internal/application/messages"
	"xadrez/internal/domain/chess"
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
	size       rl.Vector2
	client     *Client
	config     config.Config
	configPath string
	user       string
	board      *components.BoardRenderer
	invite     *messages.InviteResponse
}

func NewSession(client *Client, cfg config.Config, configPath string) *Session {
	s := &Session{client: client, config: cfg, configPath: configPath}
	s.Resize()
	return s
}

func (s *Session) Resize() {
	s.size = rl.Vector2{X: float32(rl.GetScreenWidth()), Y: float32(rl.GetScreenHeight())}
}

func (s *Session) saveConfig(cfg config.Config) error {
	if err := config.Save(s.configPath, cfg); err != nil {
		return err
	}
	s.config = cfg
	s.client.SetAddress(cfg.ServerAddress)
	return nil
}

func (s *Session) Board() *components.BoardRenderer {
	if s.board == nil {
		s.board = components.NewBoardRenderer()
	}
	return s.board
}

func (s *Session) start(user string) {
	s.user = user
	s.invite = nil
}

func (s *Session) logout() {
	s.client.Close()
	s.user = ""
	s.invite = nil
}

func (s *Session) Close() {
	if s.board != nil {
		s.board.Unload()
		s.board = nil
	}
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
	case "match_updated":
		if match, ok := event.Payload.(messages.MatchResponse); ok && s.isPlaying(match) {
			return NewMatchScreen(s, messages.RoomResponse{Room: match.Room}, match), true
		}
	}
	return nil, false
}

func (s *Session) isPlaying(match messages.MatchResponse) bool {
	return match.Outcome == chess.OutcomeNone && (match.White == s.user || match.Black == s.user)
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

func (s *Session) renderInvitePanel(x, bottom, width float32) (accepted bool) {
	if s.invite == nil {
		return false
	}
	lines := wrapText(s.invite.From+" te convidou para a sala "+s.invite.Room, textSize, int32(width-2*spacing))
	height := float32(len(lines))*float32(lineHeight) + 30 + 3*spacing
	bounds := rl.Rectangle{X: x, Y: bottom - height, Width: width, Height: height}
	rl.DrawRectangleRec(bounds, rl.NewColor(230, 230, 250, 255))

	y := bounds.Y + spacing
	for _, line := range lines {
		rl.DrawText(line, int32(x+spacing), int32(y), textSize, rl.DarkBlue)
		y += float32(lineHeight)
	}
	y += spacing

	buttonWidth := (width - 3*spacing) / 2
	if rg.Button(rl.Rectangle{X: x + spacing, Y: y, Width: buttonWidth, Height: 30}, "Aceitar") {
		accepted = true
	}
	if rg.Button(rl.Rectangle{X: x + 2*spacing + buttonWidth, Y: y, Width: buttonWidth, Height: 30}, "Recusar") {
		s.invite = nil
	}
	return accepted
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
