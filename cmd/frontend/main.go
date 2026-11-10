package main

import (
	"xadrez/cmd/frontend/components"
	"xadrez/internal/domain/chess"

	rl "github.com/gen2brain/raylib-go/raylib"
)

func main() {
	rl.InitWindow(800, 600, "Hello World")
	defer rl.CloseWindow()

	rl.SetTargetFPS(60)

	boardPieces := chess.NewGame().GetBoard()
	board := components.NewBoardRenderer(rl.Vector2{X: 800, Y: 600}, &chess.Match{
		Board: boardPieces,
	})
	defer board.Unload()

	for !rl.WindowShouldClose() {
		rl.BeginDrawing()
		rl.ClearBackground(rl.RayWhite)
		board.Render()
		rl.EndDrawing()
	}
}
