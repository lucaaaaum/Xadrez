package chess

type GameService interface {
	CreateGame(name string) (*Game, error)
	ListGames() (map[string]*Game, error)
	GetGame(id string) (*Game, error)
}
