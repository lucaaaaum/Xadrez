package main

import (
	"context"
	"xadrez/cmd/frontend/components"
	"xadrez/cmd/frontend/screens"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(800, 600, "Hello World")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	var currentScreen components.Screen = screens.NewMainMenuScreen(rl.Vector2{X: 800, Y: 600})

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)

		if currentScreen == nil {
			break
		}

		next, err := currentScreen.Update(context.Background())
		if err != nil {
			panic(err)
		}
		currentScreen.Render()
		currentScreen = next

		rl.EndDrawing()
	}
}
