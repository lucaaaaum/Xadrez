package infra

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"sync"

	"xadrez/ent"
	"xadrez/ent/user"
	"xadrez/internal/domain/users"
)

type UserService struct {
	mu     sync.Mutex
	online map[string]struct{}
	client *ent.Client
}

func NewUserService(client *ent.Client) *UserService {
	return &UserService{
		online: make(map[string]struct{}),
		client: client,
	}
}

func (s *UserService) RegisterUserLogin(ctx context.Context, u users.User) error {
	exists, err := s.client.User.Query().Where(user.Name(u.Name)).Exist(ctx)
	if err != nil {
		return fmt.Errorf("falha ao buscar o usuário: %w", err)
	}
	if !exists {
		return fmt.Errorf("o usuário %s não foi encontrado", u.Name)
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, logged := s.online[u.Name]; logged {
		return fmt.Errorf("%w: %s", users.ErrAlreadyOnline, u.Name)
	}
	s.online[u.Name] = struct{}{}
	return nil
}

func (s *UserService) RegisterUserLogout(ctx context.Context, u users.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, logged := s.online[u.Name]; !logged {
		return fmt.Errorf("o usuário %s não está logado", u.Name)
	}
	delete(s.online, u.Name)
	return nil
}

func (s *UserService) GetOnlineUsers(ctx context.Context) ([]users.User, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	list := make([]users.User, 0, len(s.online))
	for name := range s.online {
		list = append(list, users.User{Name: name})
	}
	slices.SortFunc(list, func(a, b users.User) int {
		return strings.Compare(a.Name, b.Name)
	})
	return list, nil
}
