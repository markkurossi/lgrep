//
// server.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package server

import (
	"fmt"
	"io"
	"os"

	"github.com/markkurossi/lgrep/datalog"
	"github.com/markkurossi/lgrep/syslog"
	"github.com/markkurossi/lgrep/wef"
)

// Server implements LGrep server.
type Server struct {
	DB      datalog.DB
	Syslog  *syslog.Server
	WEF     *wef.Server
	queries []*Query
}

// Query implements queries that are matched against log entries.
type Query struct {
	Clause     *datalog.Clause
	Predicates datalog.Predicates
}

// New creates a new server instance.
func New(db datalog.DB) *Server {
	server := &Server{
		DB: db,
	}
	server.Syslog = syslog.New(server)
	server.WEF = wef.New(server)
	return server
}

// Verbose sets the verbose output flag.
func (s *Server) Verbose(verbose bool) {
	s.Syslog.Verbose = verbose
	s.WEF.Verbose = verbose
}

// Add adds a clause to the server's clause database.
func (s *Server) Add(clause *datalog.Clause) {
	s.DB.Add(clause)
}

// Get gets the clauses from the server's clause database. The limits
// specify the query limits.
func (s *Server) Get(atom *datalog.Atom,
	limits datalog.Predicates) []*datalog.Clause {
	return s.DB.Get(atom, limits)
}

// Sync executes the queries against the new log entries.
func (s *Server) Sync() {
	s.executeQueries()
}

// Eval evaluates the argument file. The facts are added to the
// server's clause database, queries are executed against the
// database.
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
			s.DB.Add(clause)

		case datalog.ClauseQuery:
			fmt.Printf("Query: %s%s\n", clause, clauseType)
			s.queries = append(s.queries, &Query{
				Clause: clause,
			})
		}
	}

	// Resolve all predicates, referenced by queries.
	for _, q := range s.queries {
		q.Predicates = q.Clause.Predicates(s.DB, 0)
		if true {
			fmt.Printf("%s => %s\n", q.Clause, q.Predicates)
		}
	}

	return nil
}

func (s *Server) executeQueries() {
	for _, q := range s.queries {
		result := datalog.Execute(q.Clause.Head, s.DB, q.Predicates)
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
