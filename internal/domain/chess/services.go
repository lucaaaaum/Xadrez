package chess

import "context"

type GameService interface {
	CreateGame(ctx context.Context, name string) (*Game, error)
	ListGames(ctx context.Context) (map[string]*Game, error)
	GetGame(ctx context.Context, id string) (*Game, error)
}
