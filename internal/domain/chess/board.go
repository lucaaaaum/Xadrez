package chess

import "xadrez/internal/domain"

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

func (b *board) Move(fromX, fromY, toX, toY int, promotion PieceKind) error {
	if !inBounds(fromX, fromY) || !inBounds(toX, toY) {
		return domain.NewError("coordenadas fora do tabuleiro")
	}

	p := b.getPieceAt(fromX, fromY)
	if p == nil {
		return domain.NewError("não há peça em (%d, %d)", fromX, fromY)
	}

	target := b.getPieceAt(toX, toY)
	if target != nil && target.color == p.color {
		return domain.NewError("não é possível capturar a própria peça em (%d, %d)", toX, toY)
	}

	if !b.canMove(p, fromX, fromY, toX, toY) {
		return domain.NewError("movimento inválido para a peça em (%d, %d)", fromX, fromY)
	}

	if promotion == Pawn {
		promotion = Queen
	}
	if promotion == King {
		return domain.NewError("o peão não pode ser promovido a rei")
	}

	next := *b
	next.apply(fromX, fromY, toX, toY, promotion)
	if next.inCheck(p.color) {
		return domain.NewError("o lance deixa o próprio rei em xeque")
	}

	*b = next
	return nil
}

func (b *board) apply(fromX, fromY, toX, toY int, promotion PieceKind) {
	p := b.getPieceAt(fromX, fromY)
	b.removePiece(fromX, fromY)
	if p.kind == Pawn && (toY == 0 || toY == 7) {
		p = newPiece(promotion, p.color)
	}
	b.addPiece(p, toX, toY)
}

func (b *board) canMove(p *piece, fromX, fromY, toX, toY int) bool {
	dx, dy := toX-fromX, toY-fromY
	if dx == 0 && dy == 0 {
		return false
	}

	switch p.kind {
	case Pawn:
		return b.canMovePawn(p, fromX, fromY, toX, toY)
	case Knight:
		return (abs(dx) == 2 && abs(dy) == 1) || (abs(dx) == 1 && abs(dy) == 2)
	case Bishop:
		return abs(dx) == abs(dy) && b.pathClear(fromX, fromY, toX, toY)
	case Rook:
		return (dx == 0 || dy == 0) && b.pathClear(fromX, fromY, toX, toY)
	case Queen:
		return (abs(dx) == abs(dy) || dx == 0 || dy == 0) && b.pathClear(fromX, fromY, toX, toY)
	case King:
		return abs(dx) <= 1 && abs(dy) <= 1
	default:
		return false
	}
}

func (b *board) canMovePawn(p *piece, fromX, fromY, toX, toY int) bool {
	direction, startY := 1, 1
	if p.color == Black {
		direction, startY = -1, 6
	}
	dx, dy := toX-fromX, toY-fromY
	target := b.getPieceAt(toX, toY)

	if dx == 0 && target == nil {
		if dy == direction {
			return true
		}
		return dy == 2*direction && fromY == startY && b.getPieceAt(fromX, fromY+direction) == nil
	}
	return abs(dx) == 1 && dy == direction && target != nil && target.color != p.color
}

func (b *board) pathClear(fromX, fromY, toX, toY int) bool {
	stepX, stepY := sign(toX-fromX), sign(toY-fromY)
	for x, y := fromX+stepX, fromY+stepY; x != toX || y != toY; x, y = x+stepX, y+stepY {
		if b.getPieceAt(x, y) != nil {
			return false
		}
	}
	return true
}

func (b *board) attacks(p *piece, fromX, fromY, toX, toY int) bool {
	if p.kind == Pawn {
		direction := 1
		if p.color == Black {
			direction = -1
		}
		return abs(toX-fromX) == 1 && toY-fromY == direction
	}
	return b.canMove(p, fromX, fromY, toX, toY)
}

func (b *board) isAttacked(x, y int, by PieceColor) bool {
	for fromX := range 8 {
		for fromY := range 8 {
			p := b.getPieceAt(fromX, fromY)
			if p != nil && p.color == by && b.attacks(p, fromX, fromY, x, y) {
				return true
			}
		}
	}
	return false
}

func (b *board) inCheck(color PieceColor) bool {
	for x := range 8 {
		for y := range 8 {
			p := b.getPieceAt(x, y)
			if p != nil && p.kind == King && p.color == color {
				return b.isAttacked(x, y, color.Opponent())
			}
		}
	}
	return false
}

func (b *board) hasLegalMove(color PieceColor) bool {
	for fromX := range 8 {
		for fromY := range 8 {
			p := b.getPieceAt(fromX, fromY)
			if p == nil || p.color != color {
				continue
			}
			for toX := range 8 {
				for toY := range 8 {
					target := b.getPieceAt(toX, toY)
					if target != nil && target.color == color {
						continue
					}
					if !b.canMove(p, fromX, fromY, toX, toY) {
						continue
					}
					next := *b
					next.apply(fromX, fromY, toX, toY, Queen)
					if !next.inCheck(color) {
						return true
					}
				}
			}
		}
	}
	return false
}

func inBounds(x, y int) bool {
	return x >= 0 && x < 8 && y >= 0 && y < 8
}

func sign(x int) int {
	switch {
	case x > 0:
		return 1
	case x < 0:
		return -1
	}
	return 0
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
