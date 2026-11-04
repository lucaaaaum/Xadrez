package messages

import "xadrez/internal/domain/chess"

type Piece struct {
	Kind  chess.PieceKind  `json:"kind"`
	Color chess.PieceColor `json:"color"`
	X     int              `json:"x"`
	Y     int              `json:"y"`
}

type CreateGameRequest struct {
	Name string `json:"name"`
}

type CreateGameResponse struct {
	GameID string `json:"game_id"`
}

type ListGamesRequest struct {
}

type ListGamesResponse struct {
	GameIDs []string `json:"game_ids"`
}

type GetBoardRequest struct {
	GameID string `json:"game_id"`
}

type BoardResponse struct {
	Pieces []Piece `json:"pieces"`
}

type ValidateMoveRequest struct {
	GameID string `json:"game_id"`
	FromX  int    `json:"from_x"`
	FromY  int    `json:"from_y"`
	ToX    int    `json:"to_x"`
	ToY    int    `json:"to_y"`
}

type MoveRequest struct {
	GameID string `json:"game_id"`
	FromX  int    `json:"from_x"`
	FromY  int    `json:"from_y"`
	ToX    int    `json:"to_x"`
	ToY    int    `json:"to_y"`
}
