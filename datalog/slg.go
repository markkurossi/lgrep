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
	"math"
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
	subst, ok := env[t.Symbol]
	if ok {
		return subst
	}
	return t
}

func (t *TermVariable) Unify(o Term, env Environment) Environment {
	env[t.Symbol] = o
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
	Value string
}

func NewTermConstant(value string) Term {
	return &TermConstant{
		Value: value,
	}
}

func (t *TermConstant) Type() TermType {
	return Constant
}

func (t *TermConstant) Variable() Symbol {
	panic("Variable() called for a constant term")
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
	return t.Value
}

type Atom struct {
	Predicate Symbol
	Terms     []Term
}

func (a *Atom) String() string {
	str := a.Predicate.String()
	if len(a.Terms) > 0 {
		str += "("
		for idx, term := range a.Terms {
			if idx > 0 {
				str += ","
			}
			str += term.String()
		}
		str += ")"
	}
	return str
}

func (a *Atom) Rename(env Environment) Environment {
	for _, term := range a.Terms {
		term.Rename(env)
	}
	return env
}

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

func (a *Atom) Unify(o *Atom) Environment {
	if a.Predicate != o.Predicate {
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

func (a *Atom) Equals(o *Atom) bool {
	if a.Predicate != o.Predicate {
		return false
	}
	if len(a.Terms) != len(o.Terms) {
		return false
	}
	for idx, t := range a.Terms {
		if !t.Equals(o.Terms[idx]) {
			return false
		}
	}
	return true
}

type Clause struct {
	Head *Atom
	Body []*Atom
}

func (c *Clause) String() string {
	str := c.Head.String()
	if len(c.Body) > 0 {
		str += " :- "
		for idx, literal := range c.Body {
			if idx > 0 {
				str += ","
			}
			str += literal.String()
		}
	}
	return str
}

func (c *Clause) Equals(o *Clause) bool {
	if !c.Head.Equals(o.Head) {
		return false
	}
	if len(c.Body) != len(o.Body) {
		return false
	}
	for idx, a := range c.Body {
		if !a.Equals(o.Body[idx]) {
			return false
		}
	}
	return true
}

func (c *Clause) Resolve(o *Clause) *Clause {
	if len(c.Body) == 0 {
		return nil
	}
	env := NewEnvironment()
	o.Head.Rename(env)
	env = c.Body[0].Unify(o.Head)
	if len(env) == 0 {
		return nil
	}
	newBody := make([]*Atom, len(c.Body)-1)
	for idx, t := range c.Body[1:] {
		newBody[idx] = t.Substitute(env)
	}
	return &Clause{
		Head: c.Head.Substitute(env),
		Body: newBody,
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
			str += ","
		}
		str += fmt.Sprintf("%s=%s", k, v)
	}
	return "[" + str + "]"
}

func (c *Clause) Rename() *Clause {
	env := NewEnvironment()
	for _, atom := range c.Body {
		env = atom.Rename(env)
	}
	if len(env) == 0 {
		return c
	}
	return c.Substitute(env)
}

func (c *Clause) Substitute(env Environment) *Clause {
	if len(env) == 0 {
		return c
	}
	n := &Clause{
		Head: c.Head.Substitute(env),
		Body: make([]*Atom, 0, len(c.Body)),
	}
	for _, a := range c.Body {
		n.Body = append(n.Body, a.Substitute(env))
	}
	return n
}

type SLG struct {
	Count   int
	T       []*TableEntry
	S       *Stack
	DFN     int
	PosLink int
	NegLink int
}

type TableEntry struct {
	A    *Atom
	Anss []*Clause
	Poss []*Pair
	Negs []*Pair
	Comp bool
}

type Pair struct {
	B *Clause
	H *Clause
}

type Stack struct {
	Data []StackEntry
}

func (s *Stack) Push(a *Atom, dfn, posLink, negLink int) {
	s.Data = append(s.Data, StackEntry{
		Subgoal: a,
		DFN:     dfn,
		PosLink: posLink,
		NegLink: negLink,
	})
}

type StackEntry struct {
	Subgoal *Atom
	DFN     int
	PosLink int
	NegLink int
}

func (slg *SLG) Query(a *Atom) []*Clause {
	slg.Count = 1
	slg.T = []*TableEntry{
		&TableEntry{
			A:    a,
			Comp: false,
		},
	}
	slg.S = &Stack{}
	slg.DFN = slg.Count
	slg.PosLink = slg.DFN
	slg.NegLink = math.MaxInt32
	slg.S.Push(a, slg.DFN, slg.PosLink, slg.NegLink)
	slg.Count++
	posMin := slg.DFN
	negMin := math.MaxInt32
	slg.Subgoal(a, posMin, negMin)

	var result []*Clause
	for _, e := range slg.T {
		result = append(result, e.Anss...)
	}
	return result
}

func (slg *SLG) Subgoal(a *Atom, posMin, negMin int) {
	// for each SLG resolvent G of A :- A with some clause C e Ka
	clauses := DBClauses(a.Predicate)
	for _, clause := range clauses {
		renamed := clause.Rename()
		env := a.Unify(renamed.Head)
		if env != nil {
			substituted := renamed.Substitute(env)
			slg.NewClause(a, substituted, posMin, negMin)
		}
	}
	slg.Complete(a, posMin, negMin)
}

func (slg *SLG) NewClause(a *Atom, g *Clause, posMin, negMin int) {
	if len(g.Body) == 0 {
		slg.Answer(a, g, posMin, negMin)
	} else {
		slg.Positive(a, g, g.Body[0], posMin, negMin)
	}
}

func (slg *SLG) Complete(a *Atom, posMin, negMin int) {
	fmt.Printf("SLG.Complete(%s,%d,%d)\n", a, posMin, negMin)
	//panic("SLG.Complete()")
}

func (slg *SLG) Answer(a *Atom, g *Clause, posMin, negMin int) {
	fmt.Printf("SLG.Answer(%s,%s,%d,%d)\n", a, g, posMin, negMin)
	// XXX check if answers could be stored directly to `a'.
	var tableEntry *TableEntry
	for _, t := range slg.T {
		if a.Equals(t.A) {
			tableEntry = t
			break
		}
	}
	if tableEntry != nil {
		for _, ans := range tableEntry.Anss {
			if g.Equals(ans) {
				// Answer already added.
				return
			}
		}
	}
	tableEntry.Anss = append(tableEntry.Anss, g)

	// XXX if G has no delayed literals
	for _, pair := range tableEntry.Poss {
		resolvent := pair.H.Resolve(g)
		if resolvent != nil {
			slg.NewClause(pair.B.Head, resolvent, posMin, negMin)
		}
	}
}

func (slg *SLG) Positive(a *Atom, g *Clause, b *Atom, posMin, negMin int) {
	fmt.Printf("SLG.Positive(%s,%s,%d,%d)\n", a, g, posMin, negMin)
	//panic("SLG.Positive()")
}
