package handlers

import (
	"context"
	"slices"
	"xadrez/internal/application/messages"
	"xadrez/internal/domain/rooms"
	"xadrez/internal/domain/users"
	"xadrez/internal/infra/tcp"

	"charm.land/log/v2"
)

type Notifier interface {
	SendToUser(user, messageType string, payload any) error
}

type RoomHandler struct {
	roomService rooms.RoomService
	userService users.UserService
	notifier    Notifier
}

func NewRoomHandler(roomService rooms.RoomService, userService users.UserService, notifier Notifier) *RoomHandler {
	return &RoomHandler{
		roomService: roomService,
		userService: userService,
		notifier:    notifier,
	}
}

func (h *RoomHandler) HandleCreateRoom(ctx context.Context, request messages.CreateRoomRequest) (messages.RoomResponse, error) {
	admin, err := requireUser(ctx)
	if err != nil {
		return messages.RoomResponse{}, err
	}
	room, err := h.roomService.CreateRoom(ctx, request.Room, admin)
	if err != nil {
		return messages.RoomResponse{}, err
	}
	return roomResponse(room), nil
}

func (h *RoomHandler) HandleJoinRoom(ctx context.Context, request messages.JoinRoomRequest) (messages.RoomResponse, error) {
	name, err := requireUser(ctx)
	if err != nil {
		return messages.RoomResponse{}, err
	}
	room, err := h.roomService.JoinRoom(ctx, request.Room, name)
	if err != nil {
		return messages.RoomResponse{}, err
	}
	return roomResponse(room), nil
}

func (h *RoomHandler) HandleListRooms(ctx context.Context, request messages.ListRoomsRequest) (messages.ListRoomsResponse, error) {
	if _, err := requireUser(ctx); err != nil {
		return messages.ListRoomsResponse{}, err
	}
	list, err := h.roomService.ListRooms(ctx)
	if err != nil {
		return messages.ListRoomsResponse{}, err
	}
	response := messages.ListRoomsResponse{Rooms: make([]messages.RoomResponse, 0, len(list))}
	for _, room := range list {
		response.Rooms = append(response.Rooms, roomResponse(room))
	}
	return response, nil
}

func (h *RoomHandler) HandleInviteToRoom(ctx context.Context, request messages.InviteToRoomRequest) (messages.InviteResponse, error) {
	admin, err := requireUser(ctx)
	if err != nil {
		return messages.InviteResponse{}, err
	}
	online, err := h.userService.GetOnlineUsers(ctx)
	if err != nil {
		return messages.InviteResponse{}, err
	}
	if !slices.Contains(online, users.User{Name: request.User}) {
		return messages.InviteResponse{}, tcp.NewClientError("o usuário %s não está online", request.User)
	}

	invite, err := h.roomService.InviteToRoom(ctx, request.Room, admin, request.User)
	if err != nil {
		return messages.InviteResponse{}, err
	}
	response := messages.InviteResponse{
		Room: invite.RoomName,
		From: invite.From,
		To:   invite.To,
	}
	if err := h.notifier.SendToUser(invite.To, "room_invite", response); err != nil {
		log.Warn("falha ao entregar convite", "sala", invite.RoomName, "convidado", invite.To, "erro", err)
	}
	return response, nil
}

func roomResponse(room rooms.Room) messages.RoomResponse {
	return messages.RoomResponse{
		Room:         room.Name,
		Admin:        room.Admin,
		Participants: room.Participants,
	}
}
