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

	"github.com/markkurossi/lgrep/syslog"
)

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
		fmt.Printf("%s @ %s\n   Facility : %s\n   Severity : %s\n  Timestamp : %s\n   Hostname : %s\n    Message : %s\n",
			src, time.Now(),
			event.Facility,
			event.Severity,
			event.Timestamp,
			event.Hostname,
			event.Message)
	}
}
