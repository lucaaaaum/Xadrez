package chess

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

func (c PieceColor) Opponent() PieceColor {
	if c == White {
		return Black
	}
	return White
}

func (c PieceColor) String() string {
	if c == White {
		return "brancas"
	}
	return "pretas"
}

type piece struct {
	kind  PieceKind
	color PieceColor
}

func newPiece(kind PieceKind, color PieceColor) *piece {
	return &piece{
		kind:  kind,
		color: color,
	}
}

func (p *piece) GetKind() PieceKind {
	return p.kind
}

func (p *piece) GetColor() PieceColor {
	return p.color
}
