package main

import (
	"bufio"
	"net"

	"charm.land/log/v2"
)

func main() {
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
		go handle(conn)
	}
}

func handle(conn net.Conn) {
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
