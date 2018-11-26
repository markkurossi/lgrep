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
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"github.com/markkurossi/lgrep/datalog"
	"github.com/markkurossi/lgrep/handlers"
	"github.com/markkurossi/lgrep/syslog"
)

var logHandlers = map[string]handlers.Func{
	"sshd": handlers.SSHD,
}

var db datalog.DB

type Query struct {
	Clause     *datalog.Clause
	Predicates datalog.Predicates
}

var queries []*Query

func main() {
	verbose := flag.Bool("v", false, "Verbose output")
	init := flag.String("init", "", "Init file")
	flag.Parse()

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

	db = datalog.NewMemDB()

	if len(*init) > 0 {
		err = readInit(*init)
		if err != nil {
			log.Fatalf("Failed to read init file: %s\n", err)
		}
	}

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
			fn(event, db, *verbose)
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
				handlers.Default(event, db, *verbose)
			}
		}

		// Execute queries.
		for _, q := range queries {
			result := datalog.Query(q.Clause.Head, db, q.Predicates)
			for _, r := range result {
				for k, v := range q.Predicates {
					if r.Timestamp > v {
						q.Predicates[k] = r.Timestamp
					}
				}
				fmt.Printf("%s\n", r)
			}
		}
	}
}

func readInit(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()
	parser := datalog.NewParser(file, f)
	for {
		clause, clauseType, err := parser.Parse()
		if err != nil {
			if err != io.EOF {
				return err
			}
			break
		}
		switch clauseType {
		case datalog.ClauseFact:
			db.Add(clause)

		case datalog.ClauseQuery:
			fmt.Printf("Query: %s%s\n", clause, clauseType)
			queries = append(queries, &Query{
				Clause: clause,
			})
		}
	}

	// Resolve all predicates, referenced by queries.
	for _, q := range queries {
		q.Predicates = q.Clause.Predicates(db, 0)
		if false {
			fmt.Printf("%s => %s\n", q.Clause, q.Predicates)
		}
	}

	return nil
}
