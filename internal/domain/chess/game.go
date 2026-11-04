package chess

import "fmt"

type Game struct {
	ID    string
	board *board
}

func NewGame(name string) *Game {
	return &Game{
		ID:    name,
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

func (g *Game) GetBoard() []*piece {
	return g.board.pieces
}
