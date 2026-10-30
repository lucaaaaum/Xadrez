package main

import (
	"bufio"
	"net"
	"xadrez/internal/domain/users"
	"xadrez/internal/infra"

	"charm.land/log/v2"
)

type Services struct {
	userService users.UserService
}

func main() {
	services := newServices()

	ln, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Fatalf("ocorreu um erro durante a inicialização do servidor: %v", err)
	}
	log.Infof("servidor iniciado na porta 9000...")

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Errorf("ocorreu um erro ao conectar com um cliente: %v", err)
			continue
		}
		go handle(services, conn)
	}
}

func newServices() Services {
	return Services{
		userService: infra.NewInMemoryUserService(),
	}
}

func handle(services Services, conn net.Conn) {
	defer conn.Close()
	clientRemoteAddr := conn.RemoteAddr().String()
	log.Infof("conectado ao cliente %v...", clientRemoteAddr)
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		line := scanner.Text()
		log.Infof("[%v]: %v", clientRemoteAddr, line)
	}
	if err := scanner.Err(); err != nil {
		log.Errorf("ocorreu um erro durante a leitura de uma mensagem do cliente: %v", err)
	}
}
