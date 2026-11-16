package chess

type Square struct {
	X, Y int
}

type PlacedPiece struct {
	Kind  PieceKind
	Color PieceColor
}

func LegalMoves(position [8][8]*PlacedPiece, fromX, fromY int) []Square {
	b := &board{}
	for x, column := range position {
		for y, p := range column {
			if p != nil {
				b.addPiece(newPiece(p.Kind, p.Color), x, y)
			}
		}
	}
	return b.legalMoves(fromX, fromY)
}
