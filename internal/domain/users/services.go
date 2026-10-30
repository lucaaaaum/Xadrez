package users

import "context"

type User struct {
	Name string
}

type UserService interface {
	RegisterUserLogin(context context.Context, user User) error
	RegisterUserLogout(context context.Context, user User) error
	GetOnlineUsers(context context.Context) ([]User, error)
}
