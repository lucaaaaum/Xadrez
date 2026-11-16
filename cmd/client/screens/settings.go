package screens

import (
	"context"
	"strings"
	"xadrez/cmd/client/components"

	rg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

const settingsFieldWidth float32 = 320

type SettingsScreen struct {
	session *Session
	next    components.Screen

	address        string
	editingAddress bool
	infoMessage    string
	errorMessage   string

	save bool
	back bool
}

func NewSettingsScreen(session *Session) *SettingsScreen {
	s := &SettingsScreen{
		session:        session,
		address:        session.config.ServerAddress,
		editingAddress: true,
	}
	s.next = s
	return s
}

func (s *SettingsScreen) HandleEvent(event components.Event) {}

func (s *SettingsScreen) Update(ctx context.Context) (components.Screen, error) {
	if rl.IsKeyPressed(rl.KeyEnter) {
		s.save = true
	}
	if rl.IsKeyPressed(rl.KeyEscape) {
		s.back = true
	}
	if s.save {
		s.save = false
		cfg := s.session.config
		cfg.ServerAddress = strings.TrimSpace(s.address)
		if err := s.session.saveConfig(cfg); err != nil {
			s.errorMessage = err.Error()
			s.infoMessage = ""
		} else {
			s.errorMessage = ""
			s.infoMessage = "configuração salva"
		}
	}
	if s.back {
		s.next = NewMainMenuScreen(s.session, "")
	}
	return s.next, nil
}

func (s *SettingsScreen) Render() error {
	drawTitle("Configurações")

	x := (s.session.size.X - settingsFieldWidth) / 2
	y := s.session.size.Y/2 - 2*(buttonHeight+spacing)

	rl.DrawText("Endereço do servidor (host:porta)", int32(x), int32(y-20), textSize, rl.DarkGray)
	if rg.TextBox(rl.Rectangle{X: x, Y: y, Width: settingsFieldWidth, Height: buttonHeight}, &s.address, 128, s.editingAddress) {
		s.editingAddress = !s.editingAddress
	}
	y += buttonHeight + 2*spacing

	if rg.Button(rl.Rectangle{X: x, Y: y, Width: settingsFieldWidth, Height: buttonHeight}, "Salvar") {
		s.save = true
	}
	y += buttonHeight + spacing
	if rg.Button(rl.Rectangle{X: x, Y: y, Width: settingsFieldWidth, Height: buttonHeight}, "Voltar") {
		s.back = true
	}
	y += buttonHeight + spacing

	if s.infoMessage != "" {
		rl.DrawText(s.infoMessage, int32(x), int32(y), textSize, rl.DarkGreen)
	}
	drawError(s.errorMessage, y)

	rl.DrawText("Arquivo: "+s.session.configPath, int32(margin), int32(s.session.size.Y-margin-float32(textSize)), 14, rl.Gray)
	return nil
}
