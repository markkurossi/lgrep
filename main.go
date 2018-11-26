//
// main.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package main

import (
	"flag"
	"log"

	"github.com/markkurossi/lgrep/datalog"
	"github.com/markkurossi/lgrep/server"
)

func main() {
	verbose := flag.Bool("v", false, "Verbose output")
	init := flag.String("init", "", "Init file")
	flag.Parse()

	server := server.New(datalog.NewMemDB())
	server.Verbose = *verbose

	if len(*init) > 0 {
		err := server.Eval(*init)
		if err != nil {
			log.Fatalf("Failed to read init file: %s\n", err)
		}
	}

	server.ServeSyslogUDP(":1514")
}
