package infra

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"xadrez/ent"
	"xadrez/ent/gameroom"
	"xadrez/ent/match"
	"xadrez/ent/matchmove"
	"xadrez/ent/participant"
	"xadrez/ent/user"
	"xadrez/internal/domain"
	"xadrez/internal/domain/chess"
	"xadrez/internal/domain/rooms"

	"charm.land/log/v2"
)

type loadedMatch struct {
	record *ent.Match
	room   string
	white  string
	black  string
	game   *chess.Game
}

func (m *loadedMatch) snapshot() chess.Match {
	return chess.Match{
		Room:  m.room,
		White: m.white,
		Black: m.black,
		Turn:  m.game.CurrentTurn(),
		Board: m.game.GetBoard(),
	}
}

func (m *loadedMatch) colorOf(player string) (chess.PieceColor, bool) {
	switch player {
	case m.white:
		return chess.White, true
	case m.black:
		return chess.Black, true
	}
	return 0, false
}

type MatchService struct {
	startMu     sync.Mutex
	client      *ent.Client
	roomService rooms.RoomService
}

func NewMatchService(client *ent.Client, roomService rooms.RoomService) *MatchService {
	return &MatchService{
		client:      client,
		roomService: roomService,
	}
}

func (s *MatchService) StartMatch(ctx context.Context, roomName, admin, white, black string) (chess.Match, error) {
	room, err := s.roomService.GetRoom(ctx, roomName)
	if err != nil {
		return chess.Match{}, err
	}
	if room.Admin != admin {
		return chess.Match{}, domain.NewError("apenas o administrador pode iniciar uma partida na sala %s", roomName)
	}
	if white == black {
		return chess.Match{}, domain.NewError("os jogadores branco e preto devem ser diferentes")
	}
	for _, player := range []string{white, black} {
		if !slices.Contains(room.Participants, player) {
			return chess.Match{}, domain.NewError("o usuário %s não está na sala %s", player, roomName)
		}
	}

	s.startMu.Lock()
	defer s.startMu.Unlock()

	ongoing, err := s.client.Match.Query().
		Where(match.HasRoomWith(gameroom.Name(roomName)), match.FinishedAtIsNil()).
		Exist(ctx)
	if err != nil {
		return chess.Match{}, fmt.Errorf("falha ao buscar a partida em andamento: %w", err)
	}
	if ongoing {
		return chess.Match{}, domain.NewError("já existe uma partida em andamento na sala %s", roomName)
	}

	participants, err := s.client.Participant.Query().
		Where(
			participant.HasRoomWith(gameroom.Name(roomName)),
			participant.HasUserWith(user.NameIn(white, black)),
		).
		WithUser().
		All(ctx)
	if err != nil {
		return chess.Match{}, fmt.Errorf("falha ao buscar os participantes: %w", err)
	}
	byName := make(map[string]*ent.Participant, len(participants))
	for _, p := range participants {
		byName[p.Edges.User.Name] = p
	}
	if byName[white] == nil || byName[black] == nil {
		return chess.Match{}, fmt.Errorf("participantes da sala %s não encontrados no banco", roomName)
	}

	created, err := s.client.Match.Create().
		SetRoomID(byName[white].RoomID).
		SetWhiteParticipantID(byName[white].ID).
		SetBlackParticipantID(byName[black].ID).
		Save(ctx)
	if err != nil {
		return chess.Match{}, fmt.Errorf("falha ao criar a partida: %w", err)
	}

	log.Info("partida iniciada", "sala", roomName, "brancas", white, "pretas", black)
	loaded := &loadedMatch{record: created, room: roomName, white: white, black: black, game: chess.NewGame()}
	return loaded.snapshot(), nil
}

func (s *MatchService) GetMatch(ctx context.Context, roomName string) (chess.Match, error) {
	loaded, err := loadOngoingMatch(ctx, s.client, roomName)
	if err != nil {
		return chess.Match{}, err
	}
	return loaded.snapshot(), nil
}

func (s *MatchService) MakeMove(ctx context.Context, roomName, player string, fromX, fromY, toX, toY int) (chess.Match, error) {
	tx, err := s.client.Tx(ctx)
	if err != nil {
		return chess.Match{}, fmt.Errorf("falha ao iniciar a transação: %w", err)
	}

	loaded, err := loadOngoingMatch(ctx, tx.Client(), roomName)
	if err != nil {
		return chess.Match{}, rollback(tx, err)
	}
	color, ok := loaded.colorOf(player)
	if !ok {
		return chess.Match{}, rollback(tx, domain.NewError("o usuário %s não é jogador da partida na sala %s", player, roomName))
	}
	if err := loaded.game.MakeMove(color, fromX, fromY, toX, toY, chess.Pawn); err != nil {
		return chess.Match{}, rollback(tx, err)
	}

	ply := len(loaded.record.Edges.Moves) + 1
	_, err = tx.MatchMove.Create().
		SetMatchID(loaded.record.ID).
		SetPly(ply).
		SetFromX(fromX).
		SetFromY(fromY).
		SetToX(toX).
		SetToY(toY).
		Save(ctx)
	if ent.IsConstraintError(err) {
		return chess.Match{}, rollback(tx, domain.NewError("o tabuleiro mudou antes do lance, tente novamente"))
	}
	if err != nil {
		return chess.Match{}, rollback(tx, fmt.Errorf("falha ao gravar o lance: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return chess.Match{}, fmt.Errorf("falha ao confirmar a transação: %w", err)
	}

	return loaded.snapshot(), nil
}

func loadOngoingMatch(ctx context.Context, client *ent.Client, roomName string) (*loadedMatch, error) {
	record, err := client.Match.Query().
		Where(match.HasRoomWith(gameroom.Name(roomName)), match.FinishedAtIsNil()).
		WithWhitePlayer(func(q *ent.ParticipantQuery) { q.WithUser() }).
		WithBlackPlayer(func(q *ent.ParticipantQuery) { q.WithUser() }).
		WithMoves(func(q *ent.MatchMoveQuery) { q.Order(ent.Asc(matchmove.FieldPly)) }).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, domain.NewError("não há partida em andamento na sala %s", roomName)
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar a partida: %w", err)
	}

	game := chess.NewGame()
	for _, move := range record.Edges.Moves {
		if err := game.MakeMove(game.CurrentTurn(), move.FromX, move.FromY, move.ToX, move.ToY, chess.Pawn); err != nil {
			return nil, fmt.Errorf("lance %d inválido no histórico da partida: %w", move.Ply, err)
		}
	}

	return &loadedMatch{
		record: record,
		room:   roomName,
		white:  record.Edges.WhitePlayer.Edges.User.Name,
		black:  record.Edges.BlackPlayer.Edges.User.Name,
		game:   game,
	}, nil
}
