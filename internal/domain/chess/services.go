package chess

import "context"

type Outcome string

const (
	OutcomeNone        Outcome = ""
	OutcomeCheckmate   Outcome = "checkmate"
	OutcomeStalemate   Outcome = "stalemate"
	OutcomeResignation Outcome = "resignation"
)

type Match struct {
	Room    string
	White   string
	Black   string
	Turn    PieceColor
	InCheck bool
	Outcome Outcome
	Winner  string
	Board   [8][8]*piece
}

type MatchService interface {
	StartMatch(ctx context.Context, room, admin, white, black string) (Match, error)
	GetMatch(ctx context.Context, room string) (Match, error)
	MakeMove(ctx context.Context, room, player string, fromX, fromY, toX, toY int, promotion PieceKind) (Match, error)
	Resign(ctx context.Context, room, player string) (Match, error)
}
