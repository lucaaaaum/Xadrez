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
	s.games[name] = game
	return game, nil
}

func (s *InMemoryGameService) ListGames() (map[string]*chess.Game, error) {
	return s.games, nil
}

func (s *InMemoryGameService) GetGame(name string) (*chess.Game, error) {
	game, exists := s.games[name]
	if !exists {
		return nil, fmt.Errorf("um jogo com nome %s não foi encontrado", name)
	}
	return game, nil
}
