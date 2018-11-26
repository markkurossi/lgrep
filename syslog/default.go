//
// default.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package syslog

import (
	"fmt"

	"github.com/markkurossi/lgrep/datalog"
)

func Default(e *Event, db datalog.DB, verbose bool) {
	var predicate string
	if len(e.Ident) > 0 {
		predicate = e.Ident
	} else {
		predicate = "syslog-event"
	}

	terms := EventTerms(e)
	sym, _ := datalog.Intern(predicate, true)
	clause := datalog.NewClause(datalog.NewAtom(sym, terms), nil)
	if verbose {
		fmt.Printf("%s.\n", clause)
	}
	db.Add(clause)
}
