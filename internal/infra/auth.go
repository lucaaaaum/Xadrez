package infra

import (
	"fmt"
	"uuid"
	"xadrez/internal/domain/auth"
)

type user struct {
	name     string
	password string
	token    string
}

type InMemoryAuthService struct {
	users []user
}

func NewInMemoryAuthService() InMemoryAuthService {
	return InMemoryAuthService{
		users: make([]user, 0),
	}
}

func (s InMemoryAuthService) Login(request auth.LoginRequest) (string, error) {
	for _, user := range s.users {
		if user.name == request.Name {
			if user.password == request.Password {
				return user.token, nil
			}
			return "", fmt.Errorf("a senha do usuário está incorreta")
		}
	}	
	newUser :=  user{
		name: request.Name,
		password: request.Password,
		token: "482394",
	}
	s.users = append(s.users, newUser)
	return newUser.token, nil
}
