package handlers

import (
	"context"
	"xadrez/internal/infra/tcp"
)

func requireUser(ctx context.Context) (string, error) {
	name, ok := tcp.UserFrom(ctx)
	if !ok {
		return "", tcp.NewClientError("é necessário fazer login")
	}
	return name, nil
}
