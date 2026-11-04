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

func (p *piece) GetKind() PieceKind {
	return p.kind
}

func (p *piece) GetColor() PieceColor {
	return p.color
}

func (p *piece) GetPosition() (int, int) {
	return p.x, p.y
}
