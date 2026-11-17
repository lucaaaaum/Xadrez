package screens

import (
	"errors"
	"sync"
	"xadrez/cmd/client/components"
	"xadrez/internal/application/messages"
	"xadrez/internal/infra/tcp"
)

const DisconnectedEvent = "disconnected"

var ErrNotConnected = errors.New("sem conexão com o servidor")

type Client struct {
	address        string
	handlerManager *tcp.HandlerManager
	events         chan<- components.Event

	mu         sync.Mutex
	tcp        *tcp.TCPClient
	generation uint64
	pending    map[string]uint64
}

func NewClient(address string, handlerManager *tcp.HandlerManager, events chan<- components.Event) *Client {
	return &Client{address: address, handlerManager: handlerManager, events: events, pending: make(map[string]uint64)}
}

func (c *Client) SetAddress(address string) {
	c.mu.Lock()
	changed := c.address != address
	c.address = address
	c.mu.Unlock()
	if changed {
		c.Close()
	}
}

func (c *Client) Connected() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tcp != nil
}

func (c *Client) Connect() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.tcp != nil {
		return nil
	}
	tcpClient, err := tcp.Dial(c.address, c.handlerManager)
	if err != nil {
		return err
	}
	c.tcp = tcpClient
	clear(c.pending)
	go c.watch(tcpClient)
	return nil
}

func (c *Client) watch(tcpClient *tcp.TCPClient) {
	<-tcpClient.Done()
	c.mu.Lock()
	lost := c.tcp == tcpClient
	if lost {
		c.tcp = nil
	}
	c.mu.Unlock()
	if lost {
		c.events <- components.Event{Type: DisconnectedEvent}
	}
}

func (c *Client) Close() {
	c.mu.Lock()
	tcpClient := c.tcp
	c.tcp = nil
	c.mu.Unlock()
	if tcpClient != nil {
		tcpClient.Close()
	}
}

func (c *Client) NextGeneration() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.generation++
}

func (c *Client) Settle(event components.Event) (stale bool) {
	if event.Id == "" {
		return false
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	generation, ok := c.pending[event.Id]
	delete(c.pending, event.Id)
	return ok && generation != c.generation
}

func send[T any](c *Client, messageType string, payload T) error {
	c.mu.Lock()
	tcpClient := c.tcp
	c.mu.Unlock()
	if tcpClient == nil {
		return ErrNotConnected
	}
	id, err := tcpClient.SendRequest(messageType, payload)
	if err != nil {
		return err
	}
	c.mu.Lock()
	c.pending[id] = c.generation
	c.mu.Unlock()
	return nil
}

func (c *Client) Login(request messages.LoginRequest) error {
	return send(c, "login", request)
}

func (c *Client) ListOnlineUsers() error {
	return send(c, "list_online_users", messages.ListOnlineUsersRequest{})
}

func (c *Client) ListRooms() error {
	return send(c, "list_rooms", messages.ListRoomsRequest{})
}

func (c *Client) CreateRoom(request messages.CreateRoomRequest) error {
	return send(c, "create_room", request)
}

func (c *Client) JoinRoom(request messages.JoinRoomRequest) error {
	return send(c, "join_room", request)
}

func (c *Client) InviteToRoom(request messages.InviteToRoomRequest) error {
	return send(c, "invite_to_room", request)
}

func (c *Client) LeaveRoom(request messages.LeaveRoomRequest) error {
	return send(c, "leave_room", request)
}

func (c *Client) DeclineInvite(request messages.DeclineInviteRequest) error {
	return send(c, "decline_invite", request)
}

func (c *Client) StartMatch(request messages.StartMatchRequest) error {
	return send(c, "start_match", request)
}

func (c *Client) GetMatch(request messages.GetMatchRequest) error {
	return send(c, "get_match", request)
}

func (c *Client) MakeMove(request messages.MoveRequest) error {
	return send(c, "make_move", request)
}

func (c *Client) Resign(request messages.ResignRequest) error {
	return send(c, "resign", request)
}
