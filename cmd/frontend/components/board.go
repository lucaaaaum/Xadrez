package components

import (
	"xadrez/internal/domain/chess"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	lightTile = rl.NewColor(238, 238, 210, 255)
	darkTile  = rl.NewColor(118, 150, 86, 255)
)

type sprite struct {
	dark  rl.Texture2D
	light rl.Texture2D
}

type BoardRenderer struct {
	size    rl.Vector2
	match   *chess.Match
	sprites map[chess.PieceKind]sprite
}

func NewBoardRenderer(size rl.Vector2, match *chess.Match) *BoardRenderer {
	renderer := &BoardRenderer{
		size:    size,
		match:   match,
		sprites: make(map[chess.PieceKind]sprite),
	}

	renderer.sprites[chess.Pawn] = loadSprite("assets/pawn.png")
	renderer.sprites[chess.Rook] = loadSprite("assets/rook.png")
	renderer.sprites[chess.Knight] = loadSprite("assets/knight.png")
	renderer.sprites[chess.Bishop] = loadSprite("assets/bishop.png")
	renderer.sprites[chess.Queen] = loadSprite("assets/queen.png")
	renderer.sprites[chess.King] = loadSprite("assets/king.png")

	return renderer
}

func loadSprite(path string) sprite {
	image := rl.LoadImage(path)
	defer rl.UnloadImage(image)

	dark := rl.LoadTextureFromImage(image)
	rl.ImageColorInvert(image)
	light := rl.LoadTextureFromImage(image)

	return sprite{dark: dark, light: light}
}

func (b *BoardRenderer) Unload() {
	for _, s := range b.sprites {
		rl.UnloadTexture(s.dark)
		rl.UnloadTexture(s.light)
	}
}

func (b *BoardRenderer) tileSize() float32 {
	return min(b.size.X, b.size.Y) / 8
}

func (b *BoardRenderer) Render() error {
	tile := b.tileSize()
	for x, column := range b.match.Board {
		for y, piece := range column {
			position := rl.Vector2{X: float32(x) * tile, Y: float32(7-y) * tile}

			color := lightTile
			if (x+y)%2 == 0 {
				color = darkTile
			}
			rl.DrawRectangleV(position, rl.Vector2{X: tile, Y: tile}, color)

			if piece != nil {
				b.renderPiece(piece.GetKind(), piece.GetColor(), position, tile)
			}
		}
	}

	return nil
}

func (b *BoardRenderer) renderPiece(kind chess.PieceKind, color chess.PieceColor, position rl.Vector2, tile float32) {
	s, ok := b.sprites[kind]
	if !ok || s.dark.ID == 0 {
		return
	}

	scale := tile / float32(s.dark.Width)

	if color == chess.Black {
		rl.DrawTextureEx(s.dark, position, 0, scale, rl.White)
		return
	}

	for _, offset := range []rl.Vector2{{X: -1}, {X: 1}, {Y: -1}, {Y: 1}} {
		outline := rl.Vector2{X: position.X + offset.X*scale, Y: position.Y + offset.Y*scale}
		rl.DrawTextureEx(s.dark, outline, 0, scale, rl.White)
	}
	rl.DrawTextureEx(s.light, position, 0, scale, rl.White)
}
