package chess

import "xadrez/internal/domain"

type Status int

const (
	Ongoing Status = iota
	Checkmate
	Stalemate
)

type Game struct {
	board       *board
	currentTurn PieceColor
	status      Status
}

func NewGame() *Game {
	return &Game{
		board:       newBoard(),
		currentTurn: White,
		status:      Ongoing,
	}
}

func (g *Game) MakeMove(color PieceColor, fromX, fromY, toX, toY int, promotion PieceKind) error {
	if g.status != Ongoing {
		return domain.NewError("a partida já terminou")
	}
	if color != g.currentTurn {
		return domain.NewError("não é a vez das %s", color)
	}

	if inBounds(fromX, fromY) {
		if p := g.board.getPieceAt(fromX, fromY); p != nil && p.color != color {
			return domain.NewError("a peça em (%d, %d) não pertence às %s", fromX, fromY, color)
		}
	}

	if err := g.board.Move(fromX, fromY, toX, toY, promotion); err != nil {
		return err
	}

	g.currentTurn = g.currentTurn.Opponent()
	g.updateStatus()

	return nil
}

func (g *Game) updateStatus() {
	if g.board.hasLegalMove(g.currentTurn) {
		g.status = Ongoing
		return
	}
	if g.board.inCheck(g.currentTurn) {
		g.status = Checkmate
		return
	}
	g.status = Stalemate
}

func (g *Game) CurrentTurn() PieceColor {
	return g.currentTurn
}

func (g *Game) Status() Status {
	return g.status
}

func (g *Game) InCheck() bool {
	return g.board.inCheck(g.currentTurn)
}

func (g *Game) IsPromotion(fromX, fromY, toY int) bool {
	if !inBounds(fromX, fromY) {
		return false
	}
	p := g.board.getPieceAt(fromX, fromY)
	return p != nil && p.kind == Pawn && (toY == 0 || toY == 7)
}

func (g *Game) GetBoard() [8][8]*piece {
	return g.board.pieces
}
