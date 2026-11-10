package handlers

import (
	"context"
	"xadrez/internal/application/messages"
	"xadrez/internal/domain/chess"
	"xadrez/internal/domain/rooms"

	"charm.land/log/v2"
)

const matchUpdatedMessageType = "match_updated"

type MatchHandler struct {
	matchService chess.MatchService
	roomService  rooms.RoomService
	notifier     Notifier
}

func NewMatchHandler(matchService chess.MatchService, roomService rooms.RoomService, notifier Notifier) *MatchHandler {
	return &MatchHandler{
		matchService: matchService,
		roomService:  roomService,
		notifier:     notifier,
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
	return h.respondAndNotify(ctx, admin, match), nil
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
	match, err := h.matchService.MakeMove(ctx, request.Room, player, request.FromX, request.FromY, request.ToX, request.ToY, request.Promotion)
	if err != nil {
		return messages.MatchResponse{}, err
	}
	return h.respondAndNotify(ctx, player, match), nil
}

func (h *MatchHandler) HandleResign(ctx context.Context, request messages.ResignRequest) (messages.MatchResponse, error) {
	player, err := requireUser(ctx)
	if err != nil {
		return messages.MatchResponse{}, err
	}
	match, err := h.matchService.Resign(ctx, request.Room, player)
	if err != nil {
		return messages.MatchResponse{}, err
	}
	return h.respondAndNotify(ctx, player, match), nil
}

func (h *MatchHandler) respondAndNotify(ctx context.Context, author string, match chess.Match) messages.MatchResponse {
	response := matchResponse(match)

	room, err := h.roomService.GetRoom(ctx, match.Room)
	if err != nil {
		log.Error("falha ao buscar os participantes para notificar", "sala", match.Room, "erro", err)
		return response
	}
	for _, participant := range room.Participants {
		if participant == author {
			continue
		}
		if err := h.notifier.SendToUser(participant, matchUpdatedMessageType, response); err != nil {
			log.Debug("participante não notificado", "sala", match.Room, "usuário", participant, "erro", err)
		}
	}
	return response
}

func matchResponse(match chess.Match) messages.MatchResponse {
	response := messages.MatchResponse{
		Room:    match.Room,
		White:   match.White,
		Black:   match.Black,
		Turn:    match.Turn,
		InCheck: match.InCheck,
		Outcome: match.Outcome,
		Winner:  match.Winner,
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
