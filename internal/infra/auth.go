package infra

import (
	"fmt"
	"sync"
	"uuid"

	"charm.land/log/v2"
)

type user struct {
	name     string
	password string
	token    string
}

type InMemoryAuthService struct {
	mu    sync.Mutex
	users []user
}

func NewInMemoryAuthService() *InMemoryAuthService {
	return &InMemoryAuthService{
		users: make([]user, 0),
	}
}

func (s *InMemoryAuthService) Login(name, password string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, user := range s.users {
		if user.name == name {
			if user.password == password {
				log.Info("usuário autenticado com sucesso", "usuário", name)
				return user.token, nil
			}
			return "", fmt.Errorf("a senha do usuário está incorreta")
		}
	}
	newUser := user{
		name:     name,
		password: password,
		token:    generateToken(),
	}
	s.users = append(s.users, newUser)
	log.Info("novo usuário registrado", "usuário", name)
	return newUser.token, nil
}

func generateToken() string {
	return uuid.NewV7().String()
}
