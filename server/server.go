//
// server.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package server

import (
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"github.com/markkurossi/lgrep/datalog"
	"github.com/markkurossi/lgrep/syslog"
)

type Server struct {
	Verbose        bool
	db             datalog.DB
	queries        []*Query
	SyslogHandlers map[string]syslog.Handler
}

type Query struct {
	Clause     *datalog.Clause
	Predicates datalog.Predicates
}

func New(db datalog.DB) *Server {
	return &Server{
		db: db,
		SyslogHandlers: map[string]syslog.Handler{
			"sshd": syslog.SSHD,
		},
	}
}

func (s *Server) Eval(file string) error {
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
			s.db.Add(clause)

		case datalog.ClauseQuery:
			fmt.Printf("Query: %s%s\n", clause, clauseType)
			s.queries = append(s.queries, &Query{
				Clause: clause,
			})
		}
	}

	// Resolve all predicates, referenced by queries.
	for _, q := range s.queries {
		q.Predicates = q.Clause.Predicates(s.db, 0)
		if false {
			fmt.Printf("%s => %s\n", q.Clause, q.Predicates)
		}
	}

	return nil
}

func (s *Server) ServeSyslogUDP(address string) error {
	addr, err := net.ResolveUDPAddr("udp", address)
	if err != nil {
		return err
	}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return err
	}
	defer conn.Close()
	log.Printf("Listening at %s\n", addr)

	var buf [1024]byte
	for {
		n, _, err := conn.ReadFromUDP(buf[:])
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
		fn, ok := s.SyslogHandlers[event.Ident]
		if ok {
			fn(event, s.db, s.Verbose)
		} else {
			syslog.Default(event, s.db, s.Verbose)
		}

		s.executeQueries()
	}
}

func (s *Server) executeQueries() {
	for _, q := range s.queries {
		result := datalog.Execute(q.Clause.Head, s.db, q.Predicates)
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
