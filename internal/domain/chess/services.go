package chess

import "context"

type Match struct {
	Room  string
	White string
	Black string
	Turn  PieceColor
	Board [8][8]*piece
}

type MatchService interface {
	StartMatch(ctx context.Context, room, admin, white, black string) (Match, error)
	GetMatch(ctx context.Context, room string) (Match, error)
	MakeMove(ctx context.Context, room, player string, fromX, fromY, toX, toY int) (Match, error)
}
