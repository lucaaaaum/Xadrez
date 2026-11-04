package chess

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
