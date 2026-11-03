package chess

import "fmt"

type PieceKind int

const (
	Pawn PieceKind = iota
	Knight
	Bishop
	Rook
	Queen
	King
)

type PieceColor int

const (
	White PieceColor = iota
	Black
)

type piece struct {
	kind  PieceKind
	color PieceColor
	x     int
	y     int
}

func newPiece(kind PieceKind, color PieceColor) *piece {
	return &piece{
		kind:  kind,
		color: color,
	}
}

type board struct {
	pieces []*piece
}

func (b *board) getPieceAt(x, y int) *piece {
	for _, p := range b.pieces {
		if p.x == x && p.y == y {
			return p
		}
	}
	return nil
}

func (b *board) addPiece(p *piece, x, y int) {
	p.x = x
	p.y = y
	b.pieces = append(b.pieces, p)
}

func (b *board) removePiece(p *piece) {
	for i, piece := range b.pieces {
		if piece == p {
			b.pieces = append(b.pieces[:i], b.pieces[i+1:]...)
			return
		}
	}
}

func newBoard() *board {
	board := &board{}

	for i := range 8 {
		board.addPiece(newPiece(Pawn, White), i, 1)
		board.addPiece(newPiece(Pawn, Black), i, 6)
	}

	board.addPiece(newPiece(Rook, White), 0, 0)
	board.addPiece(newPiece(Rook, White), 7, 0)
	board.addPiece(newPiece(Rook, Black), 0, 7)
	board.addPiece(newPiece(Rook, Black), 7, 7)

	board.addPiece(newPiece(Knight, White), 1, 0)
	board.addPiece(newPiece(Knight, White), 6, 0)
	board.addPiece(newPiece(Knight, Black), 1, 7)
	board.addPiece(newPiece(Knight, Black), 6, 7)

	board.addPiece(newPiece(Bishop, White), 2, 0)
	board.addPiece(newPiece(Bishop, White), 5, 0)
	board.addPiece(newPiece(Bishop, Black), 2, 7)
	board.addPiece(newPiece(Bishop, Black), 5, 7)

	board.addPiece(newPiece(Queen, White), 3, 0)
	board.addPiece(newPiece(Queen, Black), 3, 7)

	board.addPiece(newPiece(King, White), 4, 0)
	board.addPiece(newPiece(King, Black), 4, 7)

	return board
}

type Game struct {
	board *board
}

func NewGame() *Game {
	return &Game{
		board: newBoard(),
	}
}

func (g *Game) Move(fromX, fromY, toX, toY int) error {
	p := g.board.getPieceAt(fromX, fromY)
	if p == nil {
		return fmt.Errorf("no piece at (%d, %d)", fromX, fromY)
	}

	if !canMove(p, toX, toY) {
		return fmt.Errorf("invalid move for piece %v at (%d, %d)", p.kind, fromX, fromY)
	}

	target := g.board.getPieceAt(toX, toY)
	if target != nil && target.color == p.color {
		return fmt.Errorf("cannot capture own piece at (%d, %d)", toX, toY)
	}

	p.x = toX
	p.y = toY

	if target != nil {
		g.board.removePiece(target)
	}

	return nil
}

func canMove(p *piece, toX, toY int) bool {
	switch p.kind {
	case Pawn:
		if p.color == White {
			return (toY == p.y+1 && toX == p.x) || (toY == p.y+2 && toX == p.x && p.y == 1)
		} else {
			return (toY == p.y-1 && toX == p.x) || (toY == p.y-2 && toX == p.x && p.y == 6)
		}
	case Knight:
		return (abs(toX-p.x) == 2 && abs(toY-p.y) == 1) || (abs(toX-p.x) == 1 && abs(toY-p.y) == 2)
	case Bishop:
		return abs(toX-p.x) == abs(toY-p.y)
	case Rook:
		return toX == p.x || toY == p.y
	case Queen:
		return abs(toX-p.x) == abs(toY-p.y) || toX == p.x || toY == p.y
	case King:
		return abs(toX-p.x) <= 1 && abs(toY-p.y) <= 1
	default:
		return false
	}
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
