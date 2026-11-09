package rooms

import "context"

type Room struct {
	Name         string
	Admin        string
	Participants []string
}

type Invite struct {
	RoomName string
	From     string
	To       string
}

type RoomService interface {
	CreateRoom(ctx context.Context, name, admin string) (Room, error)
	JoinRoom(ctx context.Context, name, user string) (Room, error)
	ListRooms(ctx context.Context) ([]Room, error)
	InviteToRoom(ctx context.Context, name, admin, invitee string) (Invite, error)
}
