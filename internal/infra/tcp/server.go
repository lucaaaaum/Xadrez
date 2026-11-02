package tcp

import (
	"errors"
	"fmt"
	"net"
	"sync"

	"charm.land/log/v2"
)

type TCPServer struct {
	mu             sync.Mutex
	clients        map[*Peer]struct{}
	handlerManager *HandlerManager
}

func NewTCPServer(handlerManager *HandlerManager) *TCPServer {
	return &TCPServer{
		clients:        make(map[*Peer]struct{}),
		handlerManager: handlerManager,
	}
}

func (s *TCPServer) ListenAndServe(address string) error {
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return fmt.Errorf("erro ao iniciar o servidor: %w", err)
	}
	log.Info("servidor iniciado", "endereço", ln.Addr().String())
	return s.Serve(ln)
}

func (s *TCPServer) Serve(ln net.Listener) error {
	for {
		conn, err := ln.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return nil
			}
			log.Error("erro ao aceitar conexão", "erro", err)
			continue
		}
		peer := NewPeer(conn, s.handlerManager)
		s.add(peer)
		go func() {
			peer.Run()
			s.remove(peer)
		}()
	}
}

func (s *TCPServer) Broadcast[T any](messageType string, payload T) {
	s.mu.Lock()
	clients := make([]*Peer, 0, len(s.clients))
	for client := range s.clients {
		clients = append(clients, client)
	}
	s.mu.Unlock()

	for _, client := range clients {
		if err := client.SendMessage(messageType, payload); err != nil {
			log.Error("erro ao enviar broadcast", "cliente", client.RemoteAddr(), "erro", err)
		}
	}
}

func (s *TCPServer) add(peer *Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.clients[peer] = struct{}{}
	log.Info("cliente conectado", "cliente", peer.RemoteAddr())
}

func (s *TCPServer) remove(peer *Peer) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.clients, peer)
	log.Info("cliente desconectado", "cliente", peer.RemoteAddr())
}
