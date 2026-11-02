package tcp

import (
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
)

type TCPClient struct {
	server *Peer
	nextId atomic.Uint64
}

func Dial(address string, handlerManager *HandlerManager) (*TCPClient, error) {
	conn, err := net.Dial("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("erro ao conectar ao servidor: %w", err)
	}

	peer := NewPeer(conn, handlerManager)
	client := &TCPClient{
		server: peer,
	}

	go peer.Run()

	return client, nil
}

func (c *TCPClient) SendRequest[T any](messageType string, payload T) (string, error) {
	id := strconv.FormatUint(c.nextId.Add(1), 10)
	return id, c.server.Reply(id, messageType, payload)
}

func (c *TCPClient) Close() {
	c.server.Close()
}

func (c *TCPClient) Done() <-chan struct{} {
	return c.server.Done()
}
