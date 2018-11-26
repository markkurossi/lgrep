//
// datalog.go
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

func EventTerms(e *Event) []datalog.Term {
	var terms []datalog.Term

	terms = append(terms, shared(fmt.Sprintf("%s", e.Facility), false))
	terms = append(terms, shared(fmt.Sprintf("%s", e.Severity), false))
	terms = append(terms, shared(fmt.Sprintf("%d", e.Timestamp.Unix()), false))
	terms = append(terms, datalog.NewTermConstant(e.Hostname, true))
	terms = append(terms, datalog.NewTermConstant(e.Ident, true))
	terms = append(terms, shared(fmt.Sprintf("%d", e.Pid), false))
	terms = append(terms, datalog.NewTermConstant(e.Message, true))

	return terms
}

func shared(val string, stringlike bool) datalog.Term {
	_, str := datalog.Intern(val, stringlike)
	return datalog.NewTermConstant(str, stringlike)
}
