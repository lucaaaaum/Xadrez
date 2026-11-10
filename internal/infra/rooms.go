package infra

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"xadrez/ent"
	"xadrez/ent/gameroom"
	"xadrez/ent/participant"
	"xadrez/ent/user"
	"xadrez/internal/domain"
	"xadrez/internal/domain/rooms"

	"charm.land/log/v2"
	"github.com/google/uuid"
)

type RoomService struct {
	mu      sync.Mutex
	invites map[string]map[string]rooms.Invite
	client  *ent.Client
}

func NewRoomService(client *ent.Client) *RoomService {
	return &RoomService{
		invites: make(map[string]map[string]rooms.Invite),
		client:  client,
	}
}

func (s *RoomService) CreateRoom(ctx context.Context, name, admin string) (rooms.Room, error) {
	adminID, err := s.userID(ctx, admin)
	if err != nil {
		return rooms.Room{}, err
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return rooms.Room{}, fmt.Errorf("falha ao iniciar a transação: %w", err)
	}
	created, err := tx.GameRoom.Create().SetName(name).SetAdminID(adminID).Save(ctx)
	if ent.IsValidationError(err) {
		return rooms.Room{}, rollback(tx, domain.NewError("nome de sala inválido: use letras minúsculas, números, _ ou -"))
	}
	if ent.IsConstraintError(err) {
		return rooms.Room{}, rollback(tx, domain.NewError("já existe uma sala com o nome %s", name))
	}
	if err != nil {
		return rooms.Room{}, rollback(tx, fmt.Errorf("falha ao criar a sala: %w", err))
	}
	if _, err := tx.Participant.Create().SetRoomID(created.ID).SetUserID(adminID).Save(ctx); err != nil {
		return rooms.Room{}, rollback(tx, fmt.Errorf("falha ao adicionar o administrador à sala: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return rooms.Room{}, fmt.Errorf("falha ao confirmar a transação: %w", err)
	}

	log.Info("sala criada", "sala", name, "administrador", admin)
	return rooms.Room{Name: name, Admin: admin, Participants: []string{admin}}, nil
}

func (s *RoomService) JoinRoom(ctx context.Context, name, username string) (rooms.Room, error) {
	record, err := s.activeRoom(ctx, name)
	if err != nil {
		return rooms.Room{}, err
	}
	userID, err := s.userID(ctx, username)
	if err != nil {
		return rooms.Room{}, err
	}

	_, err = s.client.Participant.Create().SetRoomID(record.ID).SetUserID(userID).Save(ctx)
	if ent.IsConstraintError(err) {
		return rooms.Room{}, domain.NewError("o usuário %s já está na sala %s", username, name)
	}
	if err != nil {
		return rooms.Room{}, fmt.Errorf("falha ao adicionar o usuário à sala: %w", err)
	}

	s.mu.Lock()
	delete(s.invites[name], username)
	s.mu.Unlock()

	log.Info("usuário entrou na sala", "sala", name, "usuário", username)
	return s.GetRoom(ctx, name)
}

func (s *RoomService) GetRoom(ctx context.Context, name string) (rooms.Room, error) {
	record, err := s.activeRoom(ctx, name)
	if err != nil {
		return rooms.Room{}, err
	}
	return toRoom(record), nil
}

func (s *RoomService) ListRooms(ctx context.Context) ([]rooms.Room, error) {
	records, err := s.roomQuery().
		Where(gameroom.StatusEQ(gameroom.StatusActive)).
		Order(ent.Asc(gameroom.FieldName)).
		All(ctx)
	if err != nil {
		return nil, fmt.Errorf("falha ao listar as salas: %w", err)
	}
	list := make([]rooms.Room, 0, len(records))
	for _, record := range records {
		list = append(list, toRoom(record))
	}
	return list, nil
}

func (s *RoomService) InviteToRoom(ctx context.Context, name, admin, invitee string) (rooms.Invite, error) {
	room, err := s.GetRoom(ctx, name)
	if err != nil {
		return rooms.Invite{}, err
	}
	if room.Admin != admin {
		return rooms.Invite{}, domain.NewError("apenas o administrador pode convidar para a sala %s", name)
	}
	if slices.Contains(room.Participants, invitee) {
		return rooms.Invite{}, domain.NewError("o usuário %s já está na sala %s", invitee, name)
	}
	if _, err := s.userID(ctx, invitee); err != nil {
		return rooms.Invite{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, exists := s.invites[name][invitee]; exists {
		return rooms.Invite{}, domain.NewError("o usuário %s já foi convidado para a sala %s", invitee, name)
	}
	if s.invites[name] == nil {
		s.invites[name] = make(map[string]rooms.Invite)
	}
	invite := rooms.Invite{RoomName: name, From: admin, To: invitee}
	s.invites[name][invitee] = invite

	log.Info("convite enviado", "sala", name, "administrador", admin, "convidado", invitee)
	return invite, nil
}

func (s *RoomService) roomQuery() *ent.GameRoomQuery {
	return s.client.GameRoom.Query().
		WithAdmin().
		WithParticipants(func(q *ent.ParticipantQuery) {
			q.WithUser().Order(ent.Asc(participant.FieldCreatedAt))
		})
}

func (s *RoomService) activeRoom(ctx context.Context, name string) (*ent.GameRoom, error) {
	record, err := s.roomQuery().
		Where(gameroom.Name(name), gameroom.StatusEQ(gameroom.StatusActive)).
		Only(ctx)
	if ent.IsNotFound(err) {
		return nil, domain.NewError("a sala %s não foi encontrada", name)
	}
	if err != nil {
		return nil, fmt.Errorf("falha ao buscar a sala: %w", err)
	}
	return record, nil
}

func toRoom(record *ent.GameRoom) rooms.Room {
	participants := make([]string, 0, len(record.Edges.Participants))
	for _, p := range record.Edges.Participants {
		participants = append(participants, p.Edges.User.Name)
	}
	return rooms.Room{
		Name:         record.Name,
		Admin:        record.Edges.Admin.Name,
		Participants: participants,
	}
}

func (s *RoomService) userID(ctx context.Context, name string) (uuid.UUID, error) {
	id, err := s.client.User.Query().Where(user.Name(name)).OnlyID(ctx)
	if ent.IsNotFound(err) {
		return uuid.UUID{}, domain.NewError("o usuário %s não foi encontrado", name)
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("falha ao buscar o usuário: %w", err)
	}
	return id, nil
}
