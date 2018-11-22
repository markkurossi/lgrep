//
// main.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/markkurossi/lgrep/datalog"
	"github.com/markkurossi/lgrep/handlers"
	"github.com/markkurossi/lgrep/syslog"
)

type LogHandler func(e *syslog.Event, db datalog.DB)

var logHandlers = map[string]LogHandler{
	"sshd": handlers.SSHD,
}

func main() {
	addr, err := net.ResolveUDPAddr("udp", ":1514")
	if err != nil {
		log.Fatalf("ResolveUDPAddr: %s\n", err)
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("ListenUDP: %s\n", err)
	}
	defer conn.Close()
	log.Printf("Listening at %s\n", addr)

	db := datalog.NewMemDB()

	var buf [1024]byte
	for {
		n, src, err := conn.ReadFromUDP(buf[:])
		if err != nil {
			log.Printf("ReadFromUDP: %s\n", err)
			continue
		}
		event, err := syslog.Parse(buf[:n])
		if err != nil {
			log.Printf("Failed to parse syslog event: %s\n%s", err,
				hex.Dump(buf[:n]))
			continue
		}
		fn, ok := logHandlers[event.Ident]
		if ok {
			fn(event, db)
		} else {
			if false {
				fmt.Printf("%s @ %s\n   Facility : %s\n   Severity : %s\n  Timestamp : %s\n   Hostname : %s\n      Ident : %s\n        Pid : %d\n    Message : %s\n",
					src, time.Now(),
					event.Facility,
					event.Severity,
					event.Timestamp,
					event.Hostname,
					event.Ident,
					event.Pid,
					event.Message)
			} else {
				handlers.Default(event, db)
			}
		}
	}
}
