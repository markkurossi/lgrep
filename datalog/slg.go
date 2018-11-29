//
// slg.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"fmt"
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
// Example 2.1. Consider a small cyclic graph and the common
// definition of transitive closure:
//
//   e(a,b).
//   e(b,c).
//   e(b,a).
//   tc(X,Y) :- e(X,Y).
//   tc(X,Y) :- e(X,Z),tc(Z,Y).
//
//   Search forest for tc(a,V):
//
//              tc(a,V) :- tc(a,V)
//                  |          |
//           +------+          +------+
//           |                        |
//           v                        v
//  tc(a,V) :- e(a,V).       tc(a,V) :- e(a,W),tc(W,V).
//           |                        |
//           v                        v
//         tc(a,b)           tc(a,V) :- tc(b,V).
//                                    |
//                           +--------+--------+
//                           |        |        |
//                           v        v        v
//                         tc(a,c). tc(a,a). tc(a,b).

func (a *Atom) RenameVariablesSLG(env EnvironmentSLG) EnvironmentSLG {
	for _, term := range a.Terms {
		term.RenameSLG(env)
	}
	return env
}

func (a *Atom) RenameSLG() *Atom {
	return a.SubstituteSLG(a.RenameVariablesSLG(NewEnvironmentSLG()))
}

func (a *Atom) SubstituteSLG(env EnvironmentSLG) *Atom {
	if len(env) == 0 {
		return a
	}
	n := &Atom{
		Predicate: a.Predicate,
		Terms:     make([]Term, 0, len(a.Terms)),
	}
	for _, t := range a.Terms {
		n.Terms = append(n.Terms, t.SubstituteSLG(env))
	}
	return n
}

func (a *Atom) UnifySLG(o *Atom) EnvironmentSLG {
	if a.Predicate != o.Predicate {
		return nil
	}
	if len(a.Terms) != len(o.Terms) {
		return nil
	}
	env := NewEnvironmentSLG()
	for i, t := range a.Terms {
		tn := t.SubstituteSLG(env)
		on := o.Terms[i].SubstituteSLG(env)
		if !tn.Equals(on) {
			env = tn.UnifySLG(on, env)
			if env == nil {
				return nil
			}
		}
	}
	return env
}

func (c *Clause) Resolve(a *Clause) *Clause {
	if len(c.Body) == 0 {
		return nil
	}
	renamed := a.Head.RenameSLG()
	env := c.Body[0].UnifySLG(renamed)
	if env == nil {
		return nil
	}
	newBody := make([]*Atom, len(c.Body)-1)
	for idx, t := range c.Body[1:] {
		newBody[idx] = t.SubstituteSLG(env)
	}
	var timestamp = a.Timestamp
	if c.Timestamp > timestamp {
		timestamp = c.Timestamp
	}
	return &Clause{
		Timestamp: timestamp,
		Head:      c.Head.SubstituteSLG(env),
		Body:      newBody,
	}
}

type EnvironmentSLG map[Symbol]Term

func NewEnvironmentSLG() EnvironmentSLG {
	return make(map[Symbol]Term)
}

func (e EnvironmentSLG) String() string {
	var str string
	for k, v := range e {
		if len(str) > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%s=%s", k, v)
	}
	return "[" + str + "]"
}

func (c *Clause) RenameSLG() *Clause {
	env := NewEnvironmentSLG()
	for _, atom := range c.Body {
		env = atom.RenameVariablesSLG(env)
	}
	if len(env) == 0 {
		return c
	}
	return c.SubstituteSLG(env)
}

func (c *Clause) SubstituteSLG(env EnvironmentSLG) *Clause {
	if len(env) == 0 {
		return c
	}
	n := &Clause{
		Timestamp: c.Timestamp,
		Head:      c.Head.SubstituteSLG(env),
		Body:      make([]*Atom, 0, len(c.Body)),
	}
	for _, a := range c.Body {
		n.Body = append(n.Body, a.SubstituteSLG(env))
	}
	return n
}

type Goals struct {
	db      DB
	limits  Predicates
	entries []*Subgoal
}

func (g *Goals) Add(entry *Subgoal) {
	e := g.Lookup(entry.Atom)
	if e == nil {
		g.entries = append(g.entries, entry)
	} else {
		e.Atom = entry.Atom
		e.Facts = entry.Facts
		e.Waiters = entry.Waiters
	}
}

func (g *Goals) Lookup(a *Atom) *Subgoal {
	for _, entry := range g.entries {
		if entry.Atom.Equals(a) {
			return entry
		}
	}
	return nil
}

func (g *Goals) Search(sg *Subgoal) {
	clauses := g.db.Get(sg.Atom.Predicate, g.limits)
	for _, clause := range clauses {
		renamed := clause.RenameSLG()
		env := sg.Atom.UnifySLG(renamed.Head)
		if env != nil {
			substituted := renamed.SubstituteSLG(env)
			g.NewClause(sg, substituted)
		}
	}
}

func (g *Goals) NewClause(sg *Subgoal, c *Clause) {
	if len(c.Body) == 0 {
		g.Fact(sg, c)
	} else {
		g.Rule(sg, c, c.Body[0])
	}
}

func (g *Goals) Fact(sg *Subgoal, a *Clause) {
	if sg.AddFact(a) {
		for _, w := range sg.Waiters {
			resolvent := w.C.Resolve(a)
			if resolvent != nil {
				g.NewClause(w.Goal, resolvent)
			}
		}
	}
}

func (g *Goals) Rule(subgoal *Subgoal, c *Clause, selected *Atom) {
	sg := g.Lookup(selected)
	if sg != nil {
		sg.Waiters = append(sg.Waiters, Waiter{
			C:    c,
			Goal: subgoal,
		})
		for _, fact := range sg.Facts {
			resolvent := c.Resolve(fact)
			if resolvent != nil {
				g.NewClause(subgoal, resolvent)
			}
		}
	} else {
		sg := &Subgoal{
			Atom: selected,
			Waiters: []Waiter{
				Waiter{
					C:    c,
					Goal: subgoal,
				},
			},
		}
		g.Add(sg)
		g.Search(sg)
	}
}

type Subgoal struct {
	Atom    *Atom
	Facts   []*Clause
	Waiters []Waiter
}

func (s *Subgoal) AddFact(a *Clause) bool {
	for _, fact := range s.Facts {
		if fact.Head.Equals(a.Head) {
			return false
		}
	}
	s.Facts = append(s.Facts, a)
	return true
}

type Waiter struct {
	C    *Clause
	Goal *Subgoal
}

func Query(a *Atom, db DB, limits Predicates) []*Clause {
	goals := &Goals{
		db:     db,
		limits: limits,
	}
	sg := &Subgoal{
		Atom: a,
	}
	goals.Add(sg)
	goals.Search(sg)

	return sg.Facts
}
