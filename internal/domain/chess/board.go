package chess

import "fmt"

type board struct {
	pieces [8][8]*piece
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

func (b *board) getPieceAt(x, y int) *piece {
	return b.pieces[x][y]
}

func (b *board) addPiece(p *piece, x, y int) {
	b.pieces[x][y] = p
}

func (b *board) removePiece(x, y int) {
	b.pieces[x][y] = nil
}

func (b *board) Move(fromX, fromY, toX, toY int) error {
	if fromX < 0 || fromX >= 8 || fromY < 0 || fromY >= 8 || toX < 0 || toX >= 8 || toY < 0 || toY >= 8 {
		return fmt.Errorf("invalid board coordinates")
	}

	p := b.getPieceAt(fromX, fromY)
	if p == nil {
		return fmt.Errorf("no piece at (%d, %d)", fromX, fromY)
	}

	if !canMove(p, fromX, fromY, toX, toY) {
		return fmt.Errorf("invalid move for piece %v at (%d, %d)", p.kind, fromX, fromY)
	}

	target := b.getPieceAt(toX, toY)
	if target != nil && target.color == p.color {
		return fmt.Errorf("cannot capture own piece at (%d, %d)", toX, toY)
	}

	b.removePiece(fromX, fromY)
	b.addPiece(p, toX, toY)

	return nil
}

func canMove(p *piece, fromX, fromY, toX, toY int) bool {
	switch p.kind {
	case Pawn:
		if p.color == White {
			return (toY == fromY+1 && toX == fromX) || (toY == fromY+2 && toX == fromX && fromY == 1)
		} else {
			return (toY == fromY-1 && toX == fromX) || (toY == fromY-2 && toX == fromX && fromY == 6)
		}
	case Knight:
		return (abs(toX-fromX) == 2 && abs(toY-fromY) == 1) || (abs(toX-fromX) == 1 && abs(toY-fromY) == 2)
	case Bishop:
		return abs(toX-fromX) == abs(toY-fromY)
	case Rook:
		return toX == fromX || toY == fromY
	case Queen:
		return abs(toX-fromX) == abs(toY-fromY) || toX == fromX || toY == fromY
	case King:
		return abs(toX-fromX) <= 1 && abs(toY-fromY) <= 1
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
