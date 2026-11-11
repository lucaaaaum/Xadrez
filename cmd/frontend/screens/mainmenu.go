package screens

import (
	"context"
	"strings"
	"xadrez/cmd/frontend/components"
	"xadrez/internal/application/messages"

	rg "github.com/gen2brain/raylib-go/raygui"
	rl "github.com/gen2brain/raylib-go/raylib"
)

type MainMenuScreen struct {
	size   rl.Vector2
	client *Client
	next   components.Screen

	name            string
	password        string
	editingName     bool
	editingPassword bool
	pendingName     string
	waiting         bool
	errorMessage    string

	submit bool
	exit   bool
}

func NewMainMenuScreen(size rl.Vector2, client *Client) *MainMenuScreen {
	m := &MainMenuScreen{
		size:        size,
		client:      client,
		editingName: true,
	}
	m.next = m
	return m
}

func (m *MainMenuScreen) HandleEvent(event components.Event) {
	if message, ok := errorMessage(event); ok {
		m.errorMessage = message
		m.waiting = false
		return
	}
	switch event.Type {
	case "login_ok":
		session := &Session{size: m.size, client: m.client, user: m.pendingName}
		m.next = NewLobbyScreen(session)
	}
}

func (m *MainMenuScreen) Update(ctx context.Context) (components.Screen, error) {
	if rl.IsKeyPressed(rl.KeyTab) {
		m.editingName, m.editingPassword = !m.editingName, m.editingName
	}
	if rl.IsKeyPressed(rl.KeyEnter) {
		m.submit = true
	}

	if m.exit {
		m.next = nil
	}
	if m.submit {
		m.submit = false
		m.login()
	}
	return m.next, nil
}

func (m *MainMenuScreen) login() {
	if m.waiting {
		return
	}
	name := strings.TrimSpace(m.name)
	if name == "" || m.password == "" {
		m.errorMessage = "informe o usuário e a senha"
		return
	}
	m.errorMessage = ""
	if err := m.client.Login(messages.LoginRequest{Name: name, Password: m.password}); err != nil {
		m.errorMessage = err.Error()
		return
	}
	m.pendingName = name
	m.waiting = true
}

func (m *MainMenuScreen) Render() error {
	drawTitle("Xadrez")

	x := (m.size.X - buttonWidth) / 2
	y := m.size.Y/2 - 2*(buttonHeight+spacing)

	rl.DrawText("Usuário", int32(x), int32(y-20), textSize, rl.DarkGray)
	if rg.TextBox(rl.Rectangle{X: x, Y: y, Width: buttonWidth, Height: buttonHeight}, &m.name, 32, m.editingName) {
		m.editingName, m.editingPassword = !m.editingName, false
	}
	y += buttonHeight + 2*spacing + 10

	rl.DrawText("Senha", int32(x), int32(y-20), textSize, rl.DarkGray)
	passwordBounds := rl.Rectangle{X: x, Y: y, Width: buttonWidth, Height: buttonHeight}
	if m.editingPassword {
		if rg.TextBox(passwordBounds, &m.password, 64, true) {
			m.editingPassword = false
		}
	} else {
		masked := strings.Repeat("*", len(m.password))
		if rg.TextBox(passwordBounds, &masked, 64, false) {
			m.editingPassword, m.editingName = true, false
		}
	}
	y += buttonHeight + 2*spacing

	label := "Entrar"
	if m.waiting {
		label = "Entrando..."
	}
	if rg.Button(rl.Rectangle{X: x, Y: y, Width: buttonWidth, Height: buttonHeight}, label) {
		m.submit = true
	}
	y += buttonHeight + spacing
	if rg.Button(rl.Rectangle{X: x, Y: y, Width: buttonWidth, Height: buttonHeight}, "Sair") {
		m.exit = true
	}
	y += buttonHeight + spacing

	drawError(m.errorMessage, y)
	return nil
}
