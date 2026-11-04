package chess

type GameService interface {
	CreateGame(name string) (*Game, error)
	ListGames() ([]*Game, error)
	GetGame(id string) (*Game, error)
	MakeMove(gameID string, fromX, fromY, toX, toY int) error
}
