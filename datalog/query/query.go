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

// a(X, Y) vs. a(mtr, Y)
func Execute(q *datalog.Atom, db datalog.DB, limits datalog.Predicates) []*datalog.Clause {
	exe := &executor{
		db:     db,
		limits: limits,
	}

	env := datalog.NewBindings()
	return exe.search(q, env)
}

type executor struct {
	db     datalog.DB
	limits datalog.Predicates
}

func (exe *executor) search(q *datalog.Atom,
	bindings datalog.Bindings) []*datalog.Clause {

	var result []*datalog.Clause

	for _, clause := range exe.db.Get(q.Predicate, exe.limits) {
		env := bindings.Clone()

		if clause.IsFact() {
			unified := q.Unify(clause.Head, env)
			if unified != nil {
				result = append(result, &datalog.Clause{
					Head: unified,
				})
			}
		} else {
			// Iterate rules
			renamed := clause.Rename()

			unified := q.Unify(renamed.Head, env)
			if unified == nil {
				fmt.Printf("Can't unify clause head: Unify(%s, %s) %s\n",
					q, renamed.Head, env)
				continue
			}

			clauses := exe.rule(unified, renamed.Body[0], renamed.Body[1:], env)
			result = append(result, clauses...)
		}
	}
	return result
}

func (exe *executor) rule(head, atom *datalog.Atom, rest []*datalog.Atom,
	bindings datalog.Bindings) []*datalog.Clause {

	var result []*datalog.Clause

	for _, clause := range exe.search(atom, bindings) {
		env := bindings.Clone()

		unified := atom.Unify(clause.Head, env)

		if len(rest) == 0 {
			if unified != nil {
				// Unified is part of the solution, and env contains
				// the bindings for the rule head.  Expand head with
				// env and add to results.
				result = append(result, &datalog.Clause{
					Head: head.Substitute(env),
				})
			}
		} else {
			// Sideways information passing strategies (SIPS)
			expanded := rest[0].Substitute(env)
			clauses := exe.rule(head, expanded, rest[1:], env)
			result = append(result, clauses...)
		}
	}
	return result
}
