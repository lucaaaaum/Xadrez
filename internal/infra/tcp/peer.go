package tcp

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"sync"
	"time"

	"charm.land/log/v2"
)

var ErrPeerClosed = errors.New("a conexão foi encerrada")
var ErrUserNotConnected = errors.New("o usuário não está conectado")

const flushTimeout = time.Second

type Peer struct {
	conn           net.Conn
	handlerManager *HandlerManager
	outbox         chan []byte
	done           chan struct{}
	closeOnce      sync.Once
	ctx            context.Context
	cancel         context.CancelFunc
	mu             sync.Mutex
	user           string
}

func NewPeer(conn net.Conn, handlerManager *HandlerManager) *Peer {
	p := &Peer{
		conn:           conn,
		handlerManager: handlerManager,
		outbox:         make(chan []byte, 64),
		done:           make(chan struct{}),
	}
	p.ctx, p.cancel = context.WithCancel(context.WithValue(context.Background(), peerKey{}, p))
	go p.writeLoop()
	return p
}

func (p *Peer) RemoteAddr() string {
	return p.conn.RemoteAddr().String()
}

func (p *Peer) Run() {
	defer p.Close()
	scanner := bufio.NewScanner(p.conn)
	for scanner.Scan() {
		p.handlerManager.HandleMessage(p, scanner.Bytes())
	}
	if err := scanner.Err(); err != nil {
		select {
		case <-p.done:
		default:
			log.Error("erro ao ler mensagem", "conexão", p.RemoteAddr(), "erro", err)
		}
	}
	p.handlerManager.HandleDisconnect(p)
}

func (p *Peer) writeLoop() {
	defer p.conn.Close()
	for {
		select {
		case line := <-p.outbox:
			if _, err := p.conn.Write(line); err != nil {
				log.Error("erro ao enviar mensagem", "conexão", p.RemoteAddr(), "erro", err)
				p.Close()
				return
			}
		case <-p.done:
			p.flush()
			return
		}
	}
}

func (p *Peer) flush() {
	p.conn.SetWriteDeadline(time.Now().Add(flushTimeout))
	for {
		select {
		case line := <-p.outbox:
			if _, err := p.conn.Write(line); err != nil {
				return
			}
		default:
			return
		}
	}
}

func (p *Peer) SendMessage[T any](messageType string, payload T) error {
	return p.Reply("", messageType, payload)
}

func (p *Peer) Reply[T any](id, messageType string, payload T) error {
	jsonPayload, err := json.Marshal(Message[T]{
		Id:      id,
		Type:    messageType,
		Payload: payload,
	})
	if err != nil {
		return fmt.Errorf("erro ao serializar a mensagem: %w", err)
	}
	return p.enqueue(append(jsonPayload, '\n'))
}

func (p *Peer) enqueue(line []byte) error {
	select {
	case <-p.done:
		return ErrPeerClosed
	default:
	}
	select {
	case p.outbox <- line:
		return nil
	case <-p.done:
		return ErrPeerClosed
	default:
		log.Warn("outbox cheia, encerrando conexão", "conexão", p.RemoteAddr())
		p.Close()
		return ErrPeerClosed
	}
}

func (p *Peer) Close() {
	p.closeOnce.Do(func() {
		close(p.done)
		p.cancel()
	})
}

func (p *Peer) Done() <-chan struct{} {
	return p.done
}

func (p *Peer) currentUser() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.user
}
