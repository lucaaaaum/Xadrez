package components

import rl "github.com/gen2brain/raylib-go/raylib"
import rg "github.com/gen2brain/raylib-go/raygui"

type Button struct {
	text      string
	bounds    rl.Rectangle
	action    func()
	isPressed bool
}

func NewButton(text string, bounds rl.Rectangle, action func()) *Button {
	return &Button{
		text:   text,
		bounds: bounds,
		action: action,
	}
}

func (b *Button) IsPressed() bool {
	return b.isPressed
}

func (b *Button) Render() {
	b.isPressed = rg.Button(b.bounds, b.text)
}

type ButtonList struct {
	position rl.Vector2
	size     rl.Vector2
	buttons  []*Button
}

func NewButtonList(position rl.Vector2, size rl.Vector2) *ButtonList {
	return &ButtonList{
		position: position,
		size:     size,
	}
}

func (bl *ButtonList) AddNewButton(text string, action func()) {
	button := NewButton(text, rl.Rectangle{}, action)
	bl.buttons = append(bl.buttons, button)
}

func (bl *ButtonList) Render() {
	for i, button := range bl.buttons {
		button.bounds = rl.Rectangle{
			X:      bl.position.X,
			Y:      bl.position.Y + float32(i)*60,
			Width:  bl.size.X,
			Height: 50,
		}
		button.Render()
	}
}

func (bl *ButtonList) Update() {
	for _, button := range bl.buttons {
		if button.IsPressed() {
			button.action()
		}
	}
}
