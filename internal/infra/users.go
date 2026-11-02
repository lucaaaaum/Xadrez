package infra

import (
	"context"
	"fmt"
	"xadrez/internal/domain/users"
)

type InMemoryUserService struct {
	users []users.User
}

func NewInMemoryUserService() InMemoryUserService {
	return InMemoryUserService{
		users: make([]users.User, 0),
	}
}

func (s InMemoryUserService) RegisterUserLogin(context context.Context, user users.User) error {
	for _, loggedInUser := range s.users {
		if loggedInUser.Name == user.Name {
			return fmt.Errorf("o usuário %v já está logado", user.Name)
		}
	}
	s.users = append(s.users, user)
	return nil
}

func (s InMemoryUserService) RegisterUserLogout(context context.Context, user users.User) error {
	for _, loggedInUser := range s.users {
		if loggedInUser.Name == user.Name {
			// remove usuário
			return nil
		}
	}
	return fmt.Errorf("o usuário %v não está logado", user.Name)
}

func (s InMemoryUserService) GetOnlineUsers(context context.Context) ([]users.User, error) {
	return s.users, nil
}
