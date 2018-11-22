//
// default.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package handlers

import (
	"fmt"

	"github.com/markkurossi/lgrep/datalog"
	"github.com/markkurossi/lgrep/syslog"
)

func Default(e *syslog.Event, db datalog.DB) {
	var predicate string
	if len(e.Ident) > 0 {
		predicate = e.Ident
	} else {
		predicate = "syslog-event"
	}

	terms := EventTerms(e)
	sym, _ := datalog.Intern(predicate, true)
	clause := datalog.NewClause(datalog.NewAtom(sym, terms), nil)
	fmt.Printf("%s.\n", clause)
	db.Add(clause)
}
