package infra

import (
	"fmt"
	"sync"
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
				return user.token, nil
			}
			return "", fmt.Errorf("a senha do usuário está incorreta")
		}
	}
	newUser := user{
		name:     name,
		password: password,
		token:    "482394",
	}
	s.users = append(s.users, newUser)
	return newUser.token, nil
}
