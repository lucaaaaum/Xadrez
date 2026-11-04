package handlers

import (
	"context"
	"xadrez/internal/application/messages"
	"xadrez/internal/domain/chess"
)

type ChessHandler struct {
	gameService chess.GameService
}

func NewChessHandler(gameService chess.GameService) *ChessHandler {
	return &ChessHandler{
		gameService: gameService,
	}
}

func (h *ChessHandler) HandleCreateGameRequest(ctx context.Context, request messages.CreateGameRequest) (messages.CreateGameResponse, error) {
	game, err := h.gameService.CreateGame(request.Name)
	if err != nil {
		return messages.CreateGameResponse{}, err
	}

	return messages.CreateGameResponse{
		GameID: game.ID,
	}, nil
}

func (h *ChessHandler) HandleListGamesRequest(ctx context.Context, request messages.ListGamesRequest) (messages.ListGamesResponse, error) {
	games, err := h.gameService.ListGames()
	if err != nil {
		return messages.ListGamesResponse{}, err
	}

	gameIDs := make([]string, len(games))
	for i, game := range games {
		gameIDs[i] = game.ID
	}

	return messages.ListGamesResponse{
		GameIDs: gameIDs,
	}, nil
}

func (h *ChessHandler) HandleGetBoardRequest(ctx context.Context, request messages.GetBoardRequest) (messages.BoardResponse, error) {
	game, err := h.gameService.GetGame(request.GameID)
	if err != nil {
		return messages.BoardResponse{}, err
	}
	return getBoardResponseFromGame(game), nil
}

func getBoardResponseFromGame(game *chess.Game) messages.BoardResponse {
	pieces := game.GetBoard()
	responsePieces := [8][8]messages.Piece{}
	for x, row := range pieces {
		for y, p := range row {
			if p == nil {
				continue
			}
			responsePieces[x][y] = messages.Piece{
				Kind:  p.GetKind(),
				Color: p.GetColor(),
			}
		}
	}
	return messages.BoardResponse{
		Pieces: responsePieces,
	}
}

func (h *ChessHandler) HandleMakeMoveRequest(ctx context.Context, request messages.MoveRequest) (messages.BoardResponse, error) {
	err := h.gameService.MakeMove(request.GameID, request.Color, request.FromX, request.FromY, request.ToX, request.ToY)
	if err != nil {
		return messages.BoardResponse{}, err
	}

	game, err := h.gameService.GetGame(request.GameID)
	if err != nil {
		return messages.BoardResponse{}, err
	}

	return getBoardResponseFromGame(game), nil
}

func (h *ChessHandler) HandleValidateMoveRequest(ctx context.Context, request messages.ValidateMoveRequest) error {
	return h.gameService.MakeMove(request.GameID, request.Color, request.FromX, request.FromY, request.ToX, request.ToY)
}
