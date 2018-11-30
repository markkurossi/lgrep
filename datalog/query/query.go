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

/*********************************** New ************************************/

func ExecuteNew(q *datalog.Atom, db datalog.DB, limits datalog.Predicates) []*datalog.Clause {
	query := &Query{
		atom:     q,
		db:       db,
		limits:   limits,
		bindings: datalog.NewBindings(),
		table:    &Table{},
	}
	return query.Search()
}

type Query struct {
	atom     *datalog.Atom
	db       datalog.DB
	limits   datalog.Predicates
	bindings datalog.Bindings
	table    *Table
	result   []*datalog.Clause
}

func (q *Query) Equals(o *Query) bool {
	return q.atom.Equals(o.atom)
}

func (q *Query) Search() []*datalog.Clause {
	found, entry := q.table.Add(q)
	if found {
		return entry.q.result
	}

	for _, clause := range q.db.Get(q.atom.Predicate, q.limits) {
		env := q.bindings.Clone()

		if clause.IsFact() {
			unified := q.atom.Unify(clause.Head, env)
			if unified != nil {
				r := &datalog.Clause{
					Head: unified,
				}
				q.addResult(r)
			}
		} else {
			// Iterate rules
			renamed := clause.Rename()

			unified := q.atom.Unify(renamed.Head, env)
			if unified == nil {
				continue
			}

			renamed.Substitute(env)

			clauses := q.rule(unified, renamed.Body[0], renamed.Body[1:],
				datalog.NewBindings())
			for _, clause := range clauses {
				q.addResult(clause)
			}
		}
	}

	// Notify waiters.
	for _, waiter := range entry.waiters {
		waiter.Search()
	}

	return q.result
}

func (q *Query) rule(head, atom *datalog.Atom, rest []*datalog.Atom,
	bindings datalog.Bindings) []*datalog.Clause {

	var result []*datalog.Clause

	subQuery := &Query{
		atom:     atom,
		db:       q.db,
		limits:   q.limits,
		bindings: bindings,
		table:    q.table,
	}

	for _, clause := range subQuery.Search() {
		env := bindings.Clone()

		unified := atom.Unify(clause.Head, env)

		if len(rest) == 0 {
			if unified != nil {
				// Unified is part of the solution, and env contains
				// the bindings for the rule head.  Expand head with
				// env and add to results.
				r := &datalog.Clause{
					Head: head.Substitute(env),
				}
				result = append(result, r)
			}
		} else {
			// Sideways information passing strategies (SIPS)
			expanded := rest[0].Substitute(env)
			clauses := q.rule(head, expanded, rest[1:], env)
			result = append(result, clauses...)
		}
	}
	return result
}

func (q *Query) addResult(result *datalog.Clause) {
	for _, r := range q.result {
		if r.Equals(result) {
			return
		}
	}
	q.result = append(q.result, result)
}

type Table struct {
	entries []*TableEntry
}

func (table *Table) Add(q *Query) (bool, *TableEntry) {
	for _, entry := range table.entries {
		if entry.q.Equals(q) {
			entry.waiters = append(entry.waiters, q)
			return true, entry
		}
	}
	entry := &TableEntry{
		q: q,
	}
	table.entries = append(table.entries, entry)
	return false, entry
}

type TableEntry struct {
	q       *Query
	waiters []*Query
}
