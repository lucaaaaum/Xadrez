package chess

import "fmt"

type Game struct {
	board       *board
	currentTurn PieceColor
}

func NewGame(name string) *Game {
	return &Game{
		board:       newBoard(),
		currentTurn: White,
	}
}

func (g *Game) MakeMove(color PieceColor, fromX, fromY, toX, toY int) error {
	if color != g.currentTurn {
		return fmt.Errorf("não é o turno da cor %v", color)
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

func (g *Game) GetBoard() [8][8]*piece {
	return g.board.pieces
}
