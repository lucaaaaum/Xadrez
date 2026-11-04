package messages

import "xadrez/internal/domain/chess"

type Piece struct {
	Kind  chess.PieceKind  `json:"kind"`
	Color chess.PieceColor `json:"color"`
}

type CreateGameRequest struct {
	Name string `json:"name"`
}

type CreateGameResponse struct {
}

type ListGamesRequest struct {
}

type ListGamesResponse struct {
	GameNames []string `json:"game_names"`
}

type GetBoardRequest struct {
	GameName string `json:"game_name"`
}

type BoardResponse struct {
	Pieces [8][8]Piece `json:"pieces"`
}

type ValidateMoveRequest struct {
	GameName string           `json:"game_name"`
	Color    chess.PieceColor `json:"color"`
	FromX    int              `json:"from_x"`
	FromY    int              `json:"from_y"`
	ToX      int              `json:"to_x"`
	ToY      int              `json:"to_y"`
}

type MoveRequest struct {
	GameName string           `json:"game_name"`
	Color    chess.PieceColor `json:"color"`
	FromX    int              `json:"from_x"`
	FromY    int              `json:"from_y"`
	ToX      int              `json:"to_x"`
	ToY      int              `json:"to_y"`
}
