package chess

import "xadrez/internal/domain"

type Game struct {
	board       *board
	currentTurn PieceColor
}

func NewGame() *Game {
	return &Game{
		board:       newBoard(),
		currentTurn: White,
	}
}

func (g *Game) MakeMove(color PieceColor, fromX, fromY, toX, toY int) error {
	if color != g.currentTurn {
		return domain.NewError("não é a vez das %s", color)
	}

	if p := g.board.getPieceAt(fromX, fromY); p != nil && p.color != color {
		return domain.NewError("a peça em (%d, %d) não pertence às %s", fromX, fromY, color)
	}

	err := g.board.Move(fromX, fromY, toX, toY)
	if err != nil {
		return err
	}

	g.switchTurn()

	return nil
}

func (g *Game) switchTurn() {
	if g.currentTurn == White {
		g.currentTurn = Black
	} else {
		g.currentTurn = White
	}
}

func (g *Game) CurrentTurn() PieceColor {
	return g.currentTurn
}

func (g *Game) GetBoard() [8][8]*piece {
	return g.board.pieces
}
