//
// query.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"fmt"
)

const debug bool = false

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

func Execute(q *Atom, db DB, limits Predicates) []*Clause {
	query := &Query{
		atom:     q,
		db:       db,
		limits:   limits,
		bindings: NewBindings(),
		table:    &Table{},
	}
	query.Search()
	return query.result
}

type Query struct {
	atom           *Atom
	db             DB
	limits         Predicates
	bindings       *Bindings
	table          *Table
	result         []*Clause
	parent         *Query
	level          int
	parentHead     *Atom
	parentRest     []*Atom
	parentBindings *Bindings
}

func (q *Query) Printf(format string, a ...interface{}) {
	for i := 0; i < q.level*4; i++ {
		fmt.Print(" ")
	}
	fmt.Printf(format, a...)
}

func (q *Query) Equals(o *Query) bool {
	return q.atom.Equals(o.atom)
}

func (q *Query) Search() {
	found, entry := q.table.Add(q)
	if found {
		q.searchResult(entry.q.result)
		return
	}

	for _, clause := range q.db.Get(q.atom.Predicate, q.limits) {
		env := q.bindings.Clone()

		if clause.IsFact() {
			if q.atom.Unify(clause.Head, env) {
				r := &Clause{
					Head: q.atom.Clone().Substitute(env),
				}
				if debug {
					q.Printf("Search.fact: %s\n", r.Head)
				}
				q.addResult(r)
			}
		} else {
			// Iterate rules
			renamed := clause.Rename()
			if !q.atom.Unify(renamed.Head, env) {
				continue
			}
			renamed.Substitute(env)

			q.rule(q.atom.Clone().Substitute(env),
				renamed.Body[0], renamed.Body[1:], NewBindings())
		}
	}

	q.searchResult(q.result)

	// Notify waiters.
	start := 0
	end := len(q.result)
	for start < end {
		for _, waiter := range entry.waiters {
			if debug {
				q.Printf("->%s %v\n", entry.q.atom, q.result[start:end])
			}
			waiter.searchResult(q.result[start:end])
		}
		start = end
		end = len(q.result)
	}
}

func (q *Query) searchResult(result []*Clause) {
	if q.parent != nil {
		q.parent.subQueryResult(q.parentHead, q.atom, q.parentRest,
			q.parentBindings, result)
	}
}

func (q *Query) rule(head, atom *Atom, rest []*Atom, bindings *Bindings) {
	subQuery := &Query{
		atom:           atom,
		db:             q.db,
		limits:         q.limits,
		bindings:       bindings,
		table:          q.table,
		parent:         q,
		level:          q.level + 1,
		parentHead:     head,
		parentRest:     rest,
		parentBindings: bindings,
	}
	subQuery.Search()
}

func (q *Query) subQueryResult(head, atom *Atom, rest []*Atom,
	bindings *Bindings, clauses []*Clause) {
	if debug {
		q.Printf("%s->%s\n", atom, clauses)
	}
	for _, clause := range clauses {
		env := bindings.Clone()

		if !atom.Unify(clause.Head, env) {
			continue
		}

		if len(rest) == 0 {
			// Unified is part of the solution, and env contains
			// the bindings for the rule head.  Expand head with
			// env and add to results.
			r := &Clause{
				Head: head.Clone().Substitute(env),
			}
			if debug {
				q.Printf("rule.fact: %s, env=%s\n", r.Head, env)
			}
			q.addResult(r)
		} else {
			// Sideways information passing strategies (SIPS)
			if debug {
				q.Printf("sips: %s\n", env)
			}
			expanded := rest[0].Clone().Substitute(env)
			q.rule(head, expanded, rest[1:], env)
		}
	}
}

func (q *Query) addResult(result *Clause) {
	if debug {
		q.Printf("%s: result %s\n", q.atom, result)
	}
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
