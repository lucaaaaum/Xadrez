package main

import (
	"net"

	"charm.land/log/v2"
)

func main() {
	_, err := net.Listen("tcp", ":9000")
	if err != nil {
		log.Error("ocorreu um erro durante a inicialização do servidor: %v", err)
	}
	log.Info("servidor iniciado na porta 9000...")
}
