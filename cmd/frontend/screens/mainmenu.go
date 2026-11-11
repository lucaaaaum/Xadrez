package screens

import (
	"context"
	"xadrez/cmd/frontend/components"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type MainMenuScreen struct {
	size    rl.Vector2
	buttons *components.ButtonList
	next    components.Screen
}

func NewMainMenuScreen(size rl.Vector2) *MainMenuScreen {
	m := &MainMenuScreen{
		size: size,
	}
	m.next = m

	buttons := components.NewButtonList(rl.Vector2{X: 20, Y: 20}, rl.Vector2{X: 100, Y: size.Y - 20})
	buttons.AddNewButton("Start Game", func() {
		// Action for starting the game
	})
	buttons.AddNewButton("Options", func() {
		// Action for opening options
	})
	buttons.AddNewButton("Exit", func() {
		m.next = nil
	})

	m.buttons = buttons

	return m
}

func (m *MainMenuScreen) Update(ctx context.Context) (components.Screen, error) {
	m.buttons.Update()
	return m.next, nil
}

func (m *MainMenuScreen) Render() error {
	m.buttons.Render()
	return nil
}
