package handlers

import (
	"context"
	"xadrez/internal/application/messages"
	"xadrez/internal/domain/users"
)

type UserHandler struct {
	userService users.UserService
}

func NewUserHandler(userService users.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

func (h *UserHandler) HandleListOnlineUsers(ctx context.Context, request messages.ListOnlineUsersRequest) (messages.ListOnlineUsersResponse, error) {
	if _, err := requireUser(ctx); err != nil {
		return messages.ListOnlineUsersResponse{}, err
	}
	online, err := h.userService.GetOnlineUsers(ctx)
	if err != nil {
		return messages.ListOnlineUsersResponse{}, err
	}
	response := messages.ListOnlineUsersResponse{Users: make([]string, 0, len(online))}
	for _, u := range online {
		response.Users = append(response.Users, u.Name)
	}
	return response, nil
}
