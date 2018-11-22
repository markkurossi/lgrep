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
	"time"
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

type TermType int

const (
	Variable TermType = iota
	Constant
)

type Term interface {
	Type() TermType
	Variable() Symbol
	Rename(env Environment)
	Substitute(env Environment) Term
	Unify(t Term, env Environment) Environment
	Equals(t Term) bool
	String() string
}

type TermVariable struct {
	Symbol Symbol
}

func NewTermVariable(symbol Symbol) Term {
	return &TermVariable{
		Symbol: symbol,
	}
}

func (t *TermVariable) Type() TermType {
	return Variable
}

func (t *TermVariable) Variable() Symbol {
	return t.Symbol
}

func (t *TermVariable) Rename(env Environment) {
	env[t.Symbol] = NewTermVariable(newUniqueSymbol())
}

func (t *TermVariable) Substitute(env Environment) Term {
	subst := env[t.Symbol]
	if subst != nil {
		return subst
	}
	return t
}

func (t *TermVariable) Unify(other Term, env Environment) Environment {
	switch o := other.(type) {
	case *TermVariable:
		env[o.Symbol] = t

	case *TermConstant:
		env[t.Symbol] = o
	}
	return env
}

func (t *TermVariable) Equals(other Term) bool {
	switch o := other.(type) {
	case *TermVariable:
		return t.Symbol == o.Symbol
	}
	return false
}

func (t *TermVariable) String() string {
	return t.Symbol.String()
}

type TermConstant struct {
	Value      string
	Stringlike bool
}

func NewTermConstant(value string, stringlike bool) Term {
	return &TermConstant{
		Value:      value,
		Stringlike: stringlike,
	}
}

func (t *TermConstant) Type() TermType {
	return Constant
}

func (t *TermConstant) Variable() Symbol {
	return NilSymbol
}

func (t *TermConstant) Rename(env Environment) {
}

func (t *TermConstant) Substitute(env Environment) Term {
	return t
}

func (t *TermConstant) Unify(other Term, env Environment) Environment {
	switch o := other.(type) {
	case *TermVariable:
		env[o.Symbol] = t
		return env

	case *TermConstant:
		if t.Value == o.Value {
			return env
		}
	}

	return nil
}

func (t *TermConstant) Equals(other Term) bool {
	switch o := other.(type) {
	case *TermConstant:
		return t.Value == o.Value
	}

	return false
}

func (t *TermConstant) String() string {
	if t.Stringlike {
		return Stringify(t.Value)
	} else {
		return t.Value
	}
}

type Atom struct {
	Predicate Symbol
	Terms     []Term
}

func NewAtom(predicate Symbol, terms []Term) *Atom {
	return &Atom{
		Predicate: predicate,
		Terms:     terms,
	}
}

func (a *Atom) String() string {
	str := a.Predicate.String()
	if len(a.Terms) > 0 {
		str += "("
		for idx, term := range a.Terms {
			if idx > 0 {
				str += ", "
			}
			str += term.String()
		}
		str += ")"
	}
	return str
}

// Rename
func (a *Atom) RenameVariables(env Environment) Environment {
	for _, term := range a.Terms {
		term.Rename(env)
	}
	return env
}
func (a *Atom) Rename() *Atom {
	return a.Substitute(a.RenameVariables(NewEnvironment()))
}

// Substitute
func (a *Atom) Substitute(env Environment) *Atom {
	if len(env) == 0 {
		return a
	}
	n := &Atom{
		Predicate: a.Predicate,
		Terms:     make([]Term, 0, len(a.Terms)),
	}
	for _, t := range a.Terms {
		n.Terms = append(n.Terms, t.Substitute(env))
	}
	return n
}

// Unify
func (a *Atom) Unify(o *Atom) Environment {
	if a.Predicate != o.Predicate {
		return nil
	}
	if len(a.Terms) != len(o.Terms) {
		return nil
	}
	env := NewEnvironment()
	for i, t := range a.Terms {
		tn := t.Substitute(env)
		on := o.Terms[i].Substitute(env)
		if !tn.Equals(on) {
			env = tn.Unify(on, env)
			if env == nil {
				return nil
			}
		}
	}
	return env
}

func (a *Atom) EqualsWithMapping(o *Atom, mapping map[Symbol]Symbol) bool {
	if a.Predicate != o.Predicate {
		return false
	}
	if len(a.Terms) != len(o.Terms) {
		return false
	}

	for idx, t := range a.Terms {
		ot := o.Terms[idx]

		switch t.Type() {
		case Variable:
			if ot.Type() != Variable {
				return false
			}
			mapped, ok := mapping[t.Variable()]
			if ok {
				if mapped != ot.Variable() {
					return false
				}
			} else {
				mapping[t.Variable()] = ot.Variable()
			}

		case Constant:
			if !t.Equals(ot) {
				return false
			}
		}
	}
	return true
}

func (a *Atom) Equals(o *Atom) bool {
	return a.EqualsWithMapping(o, make(map[Symbol]Symbol))
}

type Clause struct {
	Timestamp int64
	Head      *Atom
	Body      []*Atom
}

func (c *Clause) Fact() bool {
	return len(c.Body) == 0
}

func NewClause(head *Atom, body []*Atom) *Clause {
	return &Clause{
		Timestamp: time.Now().Unix(),
		Head:      head,
		Body:      body,
	}
}

func (c *Clause) String() string {
	str := c.Head.String()
	if len(c.Body) > 0 {
		str += " :- "
		for idx, literal := range c.Body {
			if idx > 0 {
				str += ", "
			}
			str += literal.String()
		}
	}
	return str
}

func (c *Clause) Equals(o *Clause) bool {
	mapping := make(map[Symbol]Symbol)

	if !c.Head.EqualsWithMapping(o.Head, mapping) {
		return false
	}
	if len(c.Body) != len(o.Body) {
		return false
	}
	for idx, a := range c.Body {
		if !a.EqualsWithMapping(o.Body[idx], mapping) {
			return false
		}
	}
	return true
}

func (c *Clause) Resolve(a *Clause) *Clause {
	if len(c.Body) == 0 {
		return nil
	}
	renamed := a.Head.Rename()
	env := c.Body[0].Unify(renamed)
	if env == nil {
		return nil
	}
	newBody := make([]*Atom, len(c.Body)-1)
	for idx, t := range c.Body[1:] {
		newBody[idx] = t.Substitute(env)
	}
	var timestamp = a.Timestamp
	if c.Timestamp > timestamp {
		timestamp = c.Timestamp
	}
	return &Clause{
		Timestamp: timestamp,
		Head:      c.Head.Substitute(env),
		Body:      newBody,
	}
}

type ClauseType int

const (
	ClauseError ClauseType = iota
	ClauseFact
	ClauseRetract
	ClauseQuery
)

func (t ClauseType) String() string {
	switch t {
	case ClauseError:
		return "{error}"
	case ClauseFact:
		return "."
	case ClauseRetract:
		return "~"
	case ClauseQuery:
		return "?"
	default:
		return fmt.Sprintf("{%d}", t)
	}
}

type Environment map[Symbol]Term

func NewEnvironment() Environment {
	return make(map[Symbol]Term)
}

func (e Environment) String() string {
	var str string
	for k, v := range e {
		if len(str) > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%s=%s", k, v)
	}
	return "[" + str + "]"
}

// Rename
func (c *Clause) Rename() *Clause {
	env := NewEnvironment()
	for _, atom := range c.Body {
		env = atom.RenameVariables(env)
	}
	if len(env) == 0 {
		return c
	}
	return c.Substitute(env)
}

// Substitute
func (c *Clause) Substitute(env Environment) *Clause {
	if len(env) == 0 {
		return c
	}
	n := &Clause{
		Timestamp: c.Timestamp,
		Head:      c.Head.Substitute(env),
		Body:      make([]*Atom, 0, len(c.Body)),
	}
	for _, a := range c.Body {
		n.Body = append(n.Body, a.Substitute(env))
	}
	return n
}

// Table
type Goals struct {
	db      DB
	from    int64
	entries []*Subgoal
}

// Add
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

// Search
func (g *Goals) Search(sg *Subgoal) {
	clauses := g.db.Get(sg.Atom.Predicate, g.from)
	for _, clause := range clauses {
		renamed := clause.Rename()
		env := sg.Atom.Unify(renamed.Head)
		if env != nil {
			substituted := renamed.Substitute(env)
			g.NewClause(sg, substituted)
		}
	}
}

// NewClause
func (g *Goals) NewClause(sg *Subgoal, c *Clause) {
	if len(c.Body) == 0 {
		g.Fact(sg, c)
	} else {
		g.Rule(sg, c, c.Body[0])
	}
}

// Fact
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

// Rule
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

// AddFact
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

// Query
func Query(a *Atom, db DB, from int64) []*Clause {
	goals := &Goals{
		db:   db,
		from: from,
	}
	sg := &Subgoal{
		Atom: a,
	}
	goals.Add(sg)
	goals.Search(sg)

	return sg.Facts
}
