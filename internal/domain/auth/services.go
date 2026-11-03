package auth

import "context"

type LoginService interface {
	Login(ctx context.Context, name, password string) (string, error)
}
