package handlers

import (
	"context"
	"xadrez/internal/application/messages"
	"xadrez/internal/domain/chess"
)

type MatchHandler struct {
	matchService chess.MatchService
}

func NewMatchHandler(matchService chess.MatchService) *MatchHandler {
	return &MatchHandler{
		matchService: matchService,
	}
}

func (h *MatchHandler) HandleStartMatch(ctx context.Context, request messages.StartMatchRequest) (messages.MatchResponse, error) {
	admin, err := requireUser(ctx)
	if err != nil {
		return messages.MatchResponse{}, err
	}
	match, err := h.matchService.StartMatch(ctx, request.Room, admin, request.White, request.Black)
	if err != nil {
		return messages.MatchResponse{}, err
	}
	return matchResponse(match), nil
}

func (h *MatchHandler) HandleGetMatch(ctx context.Context, request messages.GetMatchRequest) (messages.MatchResponse, error) {
	if _, err := requireUser(ctx); err != nil {
		return messages.MatchResponse{}, err
	}
	match, err := h.matchService.GetMatch(ctx, request.Room)
	if err != nil {
		return messages.MatchResponse{}, err
	}
	return matchResponse(match), nil
}

func (h *MatchHandler) HandleMakeMove(ctx context.Context, request messages.MoveRequest) (messages.MatchResponse, error) {
	player, err := requireUser(ctx)
	if err != nil {
		return messages.MatchResponse{}, err
	}
	match, err := h.matchService.MakeMove(ctx, request.Room, player, request.FromX, request.FromY, request.ToX, request.ToY)
	if err != nil {
		return messages.MatchResponse{}, err
	}
	return matchResponse(match), nil
}

func matchResponse(match chess.Match) messages.MatchResponse {
	response := messages.MatchResponse{
		Room:  match.Room,
		White: match.White,
		Black: match.Black,
		Turn:  match.Turn,
	}
	for x, row := range match.Board {
		for y, p := range row {
			if p == nil {
				continue
			}
			response.Pieces[x][y] = &messages.Piece{
				Kind:  p.GetKind(),
				Color: p.GetColor(),
			}
		}
	}
	return response
}
