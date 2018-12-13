//
// server.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package syslog

import (
	"encoding/hex"
	"log"
	"net"

	"github.com/markkurossi/lgrep/datalog"
)

type Server struct {
	Verbose  bool
	DB       datalog.DB
	Handlers map[string]Handler
}

func New(db datalog.DB) *Server {
	return &Server{
		DB: db,
		Handlers: map[string]Handler{
			"sshd": SSHD,
		},
	}
}

func (s *Server) ServeUDP(address string) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Printf("Syslog UDP: listening at %s\n", addr)

	var buf [1024]byte
	for {
		n, _, err := conn.ReadFromUDP(buf[:])
		if err != nil {
			log.Printf("ReadFromUDP: %s\n", err)
			continue
		}
		event, err := Parse(buf[:n])
		if err != nil {
			log.Printf("Failed to parse syslog event: %s\n%s", err,
				hex.Dump(buf[:n]))
			continue
		}
		fn, ok := s.Handlers[event.Ident]
		if ok {
			fn(event, s.DB, s.Verbose)
		} else {
			Default(event, s.DB, s.Verbose)
		}

		s.DB.Sync()
	}
}
