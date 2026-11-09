package messages

import "xadrez/internal/domain/chess"

type Piece struct {
	Kind  chess.PieceKind  `json:"kind"`
	Color chess.PieceColor `json:"color"`
}

type StartMatchRequest struct {
	Room  string `json:"room"`
	White string `json:"white"`
	Black string `json:"black"`
}

type GetMatchRequest struct {
	Room string `json:"room"`
}

type MoveRequest struct {
	Room  string `json:"room"`
	FromX int    `json:"from_x"`
	FromY int    `json:"from_y"`
	ToX   int    `json:"to_x"`
	ToY   int    `json:"to_y"`
}

type MatchResponse struct {
	Room   string           `json:"room"`
	White  string           `json:"white"`
	Black  string           `json:"black"`
	Turn   chess.PieceColor `json:"turn"`
	Pieces [8][8]*Piece     `json:"pieces"`
}
