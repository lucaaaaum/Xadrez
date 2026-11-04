package chess

import "fmt"

type Game struct {
	ID          string
	board       *board
	currentTurn PieceColor
}

func NewGame(name string) *Game {
	return &Game{
		ID:          name,
		board:       newBoard(),
		currentTurn: White,
	}
}

func (g *Game) MakeMove(color PieceColor, fromX, fromY, toX, toY int) error {
	if color != g.currentTurn {
		return fmt.Errorf("it's not %v's turn", color)
	}

	err := g.board.Move(fromX, fromY, toX, toY)
	if err != nil {
		return err
	}

	if g.currentTurn == White {
		g.currentTurn = Black
	} else {
		g.currentTurn = White
	}

	return nil
}

func (g *Game) GetBoard() [8][8]*piece {
	return g.board.pieces
}
