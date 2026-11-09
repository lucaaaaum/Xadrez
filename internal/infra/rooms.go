package infra

import (
	"context"
	"fmt"
	"slices"
	"sync"

	"xadrez/ent"
	"xadrez/ent/user"
	"xadrez/internal/domain/rooms"

	"charm.land/log/v2"
	"github.com/google/uuid"
)

type activeRoom struct {
	id           uuid.UUID
	name         string
	admin        string
	participants []string
	invites      map[string]rooms.Invite
}

func (r *activeRoom) snapshot() rooms.Room {
	return rooms.Room{
		Name:         r.name,
		Admin:        r.admin,
		Participants: slices.Clone(r.participants),
	}
}

type RoomService struct {
	mu     sync.Mutex
	rooms  map[string]*activeRoom
	client *ent.Client
}

func NewRoomService(client *ent.Client) *RoomService {
	return &RoomService{
		rooms:  make(map[string]*activeRoom),
		client: client,
	}
}

func (s *RoomService) CreateRoom(ctx context.Context, name, admin string) (rooms.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.rooms[name]; exists {
		return rooms.Room{}, fmt.Errorf("já existe uma sala com o nome %s", name)
	}

	adminID, err := s.userID(ctx, admin)
	if err != nil {
		return rooms.Room{}, err
	}

	tx, err := s.client.Tx(ctx)
	if err != nil {
		return rooms.Room{}, fmt.Errorf("falha ao iniciar a transação: %w", err)
	}
	created, err := tx.GameRoom.Create().SetName(name).SetAdminID(adminID).Save(ctx)
	if err != nil {
		return rooms.Room{}, rollback(tx, fmt.Errorf("falha ao criar a sala: %w", err))
	}
	if _, err := tx.Participant.Create().SetRoomID(created.ID).SetUserID(adminID).Save(ctx); err != nil {
		return rooms.Room{}, rollback(tx, fmt.Errorf("falha ao adicionar o administrador à sala: %w", err))
	}
	if err := tx.Commit(); err != nil {
		return rooms.Room{}, fmt.Errorf("falha ao confirmar a transação: %w", err)
	}

	room := &activeRoom{
		id:           created.ID,
		name:         name,
		admin:        admin,
		participants: []string{admin},
		invites:      make(map[string]rooms.Invite),
	}
	s.rooms[name] = room

	log.Info("sala criada", "sala", name, "administrador", admin)
	return room.snapshot(), nil
}

func (s *RoomService) JoinRoom(ctx context.Context, name, username string) (rooms.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, err := s.activeRoom(name)
	if err != nil {
		return rooms.Room{}, err
	}
	if slices.Contains(room.participants, username) {
		return rooms.Room{}, fmt.Errorf("o usuário %s já está na sala %s", username, name)
	}

	userID, err := s.userID(ctx, username)
	if err != nil {
		return rooms.Room{}, err
	}
	if _, err := s.client.Participant.Create().SetRoomID(room.id).SetUserID(userID).Save(ctx); err != nil {
		return rooms.Room{}, fmt.Errorf("falha ao adicionar o usuário à sala: %w", err)
	}

	room.participants = append(room.participants, username)
	delete(room.invites, username)

	log.Info("usuário entrou na sala", "sala", name, "usuário", username)
	return room.snapshot(), nil
}

func (s *RoomService) ListRooms(ctx context.Context) ([]rooms.Room, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	list := make([]rooms.Room, 0, len(s.rooms))
	for _, room := range s.rooms {
		list = append(list, room.snapshot())
	}
	return list, nil
}

func (s *RoomService) InviteToRoom(ctx context.Context, name, admin, invitee string) (rooms.Invite, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	room, err := s.activeRoom(name)
	if err != nil {
		return rooms.Invite{}, err
	}
	if room.admin != admin {
		return rooms.Invite{}, fmt.Errorf("apenas o administrador pode convidar para a sala %s", name)
	}
	if slices.Contains(room.participants, invitee) {
		return rooms.Invite{}, fmt.Errorf("o usuário %s já está na sala %s", invitee, name)
	}
	if _, exists := room.invites[invitee]; exists {
		return rooms.Invite{}, fmt.Errorf("o usuário %s já foi convidado para a sala %s", invitee, name)
	}
	if _, err := s.userID(ctx, invitee); err != nil {
		return rooms.Invite{}, err
	}

	invite := rooms.Invite{RoomName: name, From: admin, To: invitee}
	room.invites[invitee] = invite

	log.Info("convite enviado", "sala", name, "administrador", admin, "convidado", invitee)
	return invite, nil
}

func (s *RoomService) activeRoom(name string) (*activeRoom, error) {
	room, exists := s.rooms[name]
	if !exists {
		return nil, fmt.Errorf("a sala %s não foi encontrada", name)
	}
	return room, nil
}

func (s *RoomService) userID(ctx context.Context, name string) (uuid.UUID, error) {
	id, err := s.client.User.Query().Where(user.Name(name)).OnlyID(ctx)
	if ent.IsNotFound(err) {
		return uuid.UUID{}, fmt.Errorf("o usuário %s não foi encontrado", name)
	}
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("falha ao buscar o usuário: %w", err)
	}
	return id, nil
}
