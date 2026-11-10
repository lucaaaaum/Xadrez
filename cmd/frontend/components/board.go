package components

import (
	"xadrez/internal/domain/chess"

	rl "github.com/gen2brain/raylib-go/raylib"
)

type BoardRenderer struct {
	size    rl.Vector2
	match   *chess.Match
	sprites map[chess.PieceKind]rl.Texture2D
}

func NewBoardRenderer(size rl.Vector2, match *chess.Match) *BoardRenderer {
	renderer := &BoardRenderer{
		size:    size,
		match:   match,
		sprites: make(map[chess.PieceKind]rl.Texture2D),
	}

	renderer.sprites[chess.Pawn] = rl.LoadTexture("assets/pawn.png")
	renderer.sprites[chess.Rook] = rl.LoadTexture("assets/rook.png")
	renderer.sprites[chess.Knight] = rl.LoadTexture("assets/knight.png")
	renderer.sprites[chess.Bishop] = rl.LoadTexture("assets/bishop.png")
	renderer.sprites[chess.Queen] = rl.LoadTexture("assets/queen.png")
	renderer.sprites[chess.King] = rl.LoadTexture("assets/king.png")

	return renderer
}

func (b *BoardRenderer) Render() error {
	currentColor := chess.White
	tileSize := rl.Vector2{X: b.size.X / 8, Y: b.size.Y / 8}
	for x, column := range b.match.Board {
		for y, piece := range column {

			position := rl.Vector2{X: float32(x) * tileSize.X, Y: float32(y) * tileSize.Y}

			var color rl.Color
			if currentColor == chess.White {
				color = rl.RayWhite
			} else {
				color = rl.Black
			}

			rl.DrawRectangle(int32(position.X), int32(position.Y), int32(tileSize.X), int32(tileSize.Y), color)

			if piece != nil {
				b.renderPiece(piece.GetKind(), piece.GetColor(), position, tileSize)
			}

			currentColor = invertColor(currentColor)
		}

		currentColor = invertColor(currentColor)
	}

	return nil
}

func invertColor(color chess.PieceColor) chess.PieceColor {
	if color == chess.White {
		return chess.Black
	}
	return chess.White
}

func (b *BoardRenderer) renderPiece(kind chess.PieceKind, color chess.PieceColor, position rl.Vector2, tileSize rl.Vector2) {
	var pieceColor rl.Color
	if color == chess.White {
		pieceColor = rl.White
	} else {
		pieceColor = rl.Black
	}
	sprite := b.sprites[kind]
	rl.DrawTextureEx(sprite, position, 0, tileSize.X/float32(16), pieceColor)
}
