package components

import (
	"xadrez/internal/application/messages"
	"xadrez/internal/domain/chess"

	rl "github.com/gen2brain/raylib-go/raylib"
)

var (
	lightTile     = rl.NewColor(240, 160, 70, 255)
	darkTile      = rl.NewColor(120, 70, 160, 255)
	selectedTile  = rl.NewColor(255, 255, 255, 110)
	highlightTile = rl.NewColor(80, 220, 120, 110)
)

type Square struct {
	X, Y int
}

type BoardLayout struct {
	Bounds  rl.Rectangle
	Flipped bool
}

func (l BoardLayout) tileSize() float32 {
	return min(l.Bounds.Width, l.Bounds.Height) / 8
}

func (l BoardLayout) squarePosition(s Square) rl.Vector2 {
	tile := l.tileSize()
	column, row := s.X, 7-s.Y
	if l.Flipped {
		column, row = 7-s.X, s.Y
	}
	return rl.Vector2{
		X: l.Bounds.X + float32(column)*tile,
		Y: l.Bounds.Y + float32(row)*tile,
	}
}

func (l BoardLayout) SquareAt(point rl.Vector2) (Square, bool) {
	tile := l.tileSize()
	if tile <= 0 {
		return Square{}, false
	}

	column := int((point.X - l.Bounds.X) / tile)
	row := int((point.Y - l.Bounds.Y) / tile)
	if point.X < l.Bounds.X || point.Y < l.Bounds.Y || column > 7 || row > 7 {
		return Square{}, false
	}

	if l.Flipped {
		return Square{X: 7 - column, Y: row}, true
	}
	return Square{X: column, Y: 7 - row}, true
}

type BoardView struct {
	Layout     BoardLayout
	Pieces     [8][8]*messages.Piece
	Selected   *Square
	Highlights []Square
}

type sprite struct {
	dark  rl.Texture2D
	light rl.Texture2D
}

type BoardRenderer struct {
	sprites map[chess.PieceKind]sprite
}

func NewBoardRenderer() *BoardRenderer {
	return &BoardRenderer{
		sprites: map[chess.PieceKind]sprite{
			chess.Pawn:   loadSprite("assets/pawn.png"),
			chess.Rook:   loadSprite("assets/rook.png"),
			chess.Knight: loadSprite("assets/knight.png"),
			chess.Bishop: loadSprite("assets/bishop.png"),
			chess.Queen:  loadSprite("assets/queen.png"),
			chess.King:   loadSprite("assets/king.png"),
		},
	}
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

func (b *BoardRenderer) Render(view BoardView) {
	tile := view.Layout.tileSize()
	tileSize := rl.Vector2{X: tile, Y: tile}

	for x := range 8 {
		for y := range 8 {
			color := lightTile
			if (x+y)%2 == 0 {
				color = darkTile
			}
			rl.DrawRectangleV(view.Layout.squarePosition(Square{X: x, Y: y}), tileSize, color)
		}
	}

	for _, s := range view.Highlights {
		rl.DrawRectangleV(view.Layout.squarePosition(s), tileSize, highlightTile)
	}
	if view.Selected != nil {
		rl.DrawRectangleV(view.Layout.squarePosition(*view.Selected), tileSize, selectedTile)
	}

	for x, column := range view.Pieces {
		for y, piece := range column {
			if piece != nil {
				b.renderPiece(*piece, view.Layout.squarePosition(Square{X: x, Y: y}), tile)
			}
		}
	}
}

func (b *BoardRenderer) renderPiece(piece messages.Piece, position rl.Vector2, tile float32) {
	s, ok := b.sprites[piece.Kind]
	if !ok || s.dark.ID == 0 {
		return
	}

	scale := tile / float32(s.dark.Width)

	texture := s.light
	if piece.Color == chess.Black {
		texture = s.dark
	}
	rl.DrawTextureEx(texture, position, 0, scale, rl.White)
}
