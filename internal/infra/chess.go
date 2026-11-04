package infra

import (
	"fmt"
	"xadrez/internal/domain/chess"
)

type InMemoryGameService struct {
	games map[string]*chess.Game
}

func NewInMemoryGameService() *InMemoryGameService {
	return &InMemoryGameService{
		games: make(map[string]*chess.Game),
	}
}

func (s *InMemoryGameService) CreateGame(name string) (*chess.Game, error) {
	game := chess.NewGame(name)
	s.games[game.ID] = game
	return game, nil
}

func (s *InMemoryGameService) ListGames() ([]*chess.Game, error) {
	games := make([]*chess.Game, 0, len(s.games))
	for _, game := range s.games {
		games = append(games, game)
	}
	return games, nil
}

func (s *InMemoryGameService) GetGame(id string) (*chess.Game, error) {
	game, exists := s.games[id]
	if !exists {
		return nil, fmt.Errorf("game with ID %s not found", id)
	}
	return game, nil
}

func (s *InMemoryGameService) MakeMove(gameID string, color chess.PieceColor, fromX, fromY, toX, toY int) error {
	game, err := s.GetGame(gameID)
	if err != nil {
		return err
	}
	return game.MakeMove(color, fromX, fromY, toX, toY)
}
