package messages

type CreateRoomRequest struct {
	Room string `json:"room"`
}

type JoinRoomRequest struct {
	Room string `json:"room"`
}

type ListRoomsRequest struct {
}

type ListRoomsResponse struct {
	Rooms []RoomResponse `json:"rooms"`
}

type RoomResponse struct {
	Room         string   `json:"room"`
	Admin        string   `json:"admin"`
	Participants []string `json:"participants"`
}

type InviteToRoomRequest struct {
	Room string `json:"room"`
	User string `json:"user"`
}

type InviteResponse struct {
	Room string `json:"room"`
	From string `json:"from"`
	To   string `json:"to"`
}
