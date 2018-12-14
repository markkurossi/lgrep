//
// datalog.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package wef

import (
	"fmt"

	"github.com/markkurossi/lgrep/datalog"
)

func (s *Server) datalog(e *Event) {
	var terms []datalog.Term

	var fmtLevel string
	var fmtTask string
	var fmtOpcode string

	if e.RenderingInfo != nil {
		fmtLevel = e.RenderingInfo.Level
		fmtTask = e.RenderingInfo.Task
		fmtOpcode = e.RenderingInfo.Opcode
	}

	sym, _ := datalog.Intern(e.System.Provider.Name, true)

	terms = append(terms, constant(e.System.EventID, false))
	terms = append(terms, shared(e.System.Version, false))
	terms = append(terms, shared(e.System.Level, false))
	terms = append(terms, shared(fmtLevel, false))
	terms = append(terms, constant(e.System.Task, false))
	terms = append(terms, constant(fmtTask, true))
	terms = append(terms, shared(e.System.Opcode, false))
	terms = append(terms, shared(fmtOpcode, false))
	terms = append(terms, constant(e.System.Keywords, false))

	// XXX parse and store as int
	terms = append(terms, constant(e.System.TimeCreated.SystemTime, true))

	terms = append(terms, constant(e.System.EventRecordID, false))
	terms = append(terms, shared(e.System.Channel, false))
	terms = append(terms, constant(e.System.Computer, true))

	var str string
	if e.System.Security != nil {
		str = e.System.Security.UserID
	}
	terms = append(terms, constant(str, true))

	for _, ed := range e.EventData {
		terms = append(terms, constant(ed.Value, true))
	}

	clause := datalog.NewClause(datalog.NewAtom(sym, terms), nil)
	if s.Verbose {
		fmt.Printf("%s.\n", clause)
	}
	s.DB.Add(clause)
}

func shared(val string, stringlike bool) datalog.Term {
	_, str := datalog.Intern(val, stringlike)
	return datalog.NewTermConstant(str, stringlike)
}

func constant(val string, stringlike bool) datalog.Term {
	return datalog.NewTermConstant(val, stringlike)
}
