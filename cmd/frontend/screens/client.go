package screens

import (
	"xadrez/internal/application/messages"
	"xadrez/internal/infra/tcp"
)

type Client struct {
	tcp *tcp.TCPClient
}

func NewClient(tcpClient *tcp.TCPClient) *Client {
	return &Client{tcp: tcpClient}
}

func (c *Client) Login(request messages.LoginRequest) error {
	_, err := c.tcp.SendRequest("login", request)
	return err
}

func (c *Client) ListOnlineUsers() error {
	_, err := c.tcp.SendRequest("list_online_users", messages.ListOnlineUsersRequest{})
	return err
}

func (c *Client) ListRooms() error {
	_, err := c.tcp.SendRequest("list_rooms", messages.ListRoomsRequest{})
	return err
}

func (c *Client) CreateRoom(request messages.CreateRoomRequest) error {
	_, err := c.tcp.SendRequest("create_room", request)
	return err
}

func (c *Client) JoinRoom(request messages.JoinRoomRequest) error {
	_, err := c.tcp.SendRequest("join_room", request)
	return err
}

func (c *Client) InviteToRoom(request messages.InviteToRoomRequest) error {
	_, err := c.tcp.SendRequest("invite_to_room", request)
	return err
}

func (c *Client) StartMatch(request messages.StartMatchRequest) error {
	_, err := c.tcp.SendRequest("start_match", request)
	return err
}

func (c *Client) GetMatch(request messages.GetMatchRequest) error {
	_, err := c.tcp.SendRequest("get_match", request)
	return err
}

func (c *Client) MakeMove(request messages.MoveRequest) error {
	_, err := c.tcp.SendRequest("make_move", request)
	return err
}

func (c *Client) Resign(request messages.ResignRequest) error {
	_, err := c.tcp.SendRequest("resign", request)
	return err
}
