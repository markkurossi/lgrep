//
// query.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package query

import (
	"fmt"

	"github.com/markkurossi/lgrep/datalog"
)

// A `program' is a finite set of clauses of the form:
//
//   A :- L1,...,Ln
//
// Where A is an atom and L1,...,Ln are literals. When n=0, a clause,
// possibly containing variables, is called a `fact'. By a `subgoal',
// we mean an atom. Subgoals (and literals) that are variants of each
// other are considered syntactically identical.
//
// - Extensional database predicates (EDB) – source tables
// - Intensional database predicates (IDB) – derived tables
//
// Bottom-up Evaluation
// - Start with the EDB relations
// - Apply rules till convergence
// – Construct the minimal Herbrand model

// a(X, Y) vs. a(mtr, Y)
func Execute(a *datalog.Atom, db datalog.DB, limits datalog.Predicates) []*datalog.Clause {
	exe := &executor{
		db:     db,
		limits: limits,
	}
	exe.search(a)
	return exe.results
}

type executor struct {
	db      datalog.DB
	limits  datalog.Predicates
	results []*datalog.Clause
}

// XXX not sure about environment here.
func (e *executor) search(a *datalog.Atom) {
	for _, clause := range e.db.Get(a.Predicate, e.limits) {
		if clause.IsFact() {
			unified := a.Unify(clause.Head, datalog.NewEnvironment())
			if unified != nil {
				e.AddResult(&datalog.Clause{
					Head: unified,
				})
			}
		} else {
			// Iterate rules
			fmt.Printf("Rules not implemented yet\n")
		}
	}
}

func (e *executor) AddResult(clause *datalog.Clause) {
	e.results = append(e.results, clause)
}
