//
// atom.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"fmt"
)

// Flags define atom flags
type Flags int

// Known atom flags.
const (
	FlagPersistent Flags = 1 << iota
)

// Atom implements datalog atoms.
type Atom struct {
	Predicate Symbol
	Terms     []Term
	Flags     Flags
}

// AtomID defines atom IDs.
type AtomID uint64

func (id AtomID) String() string {
	return fmt.Sprintf("%s/%d", id.Symbol(), id.Arity())
}

// Symbol returns the atom symbol.
func (id AtomID) Symbol() Symbol {
	return Symbol(id >> 32)
}

// Arity returns the atom arity.
func (id AtomID) Arity() int {
	return int(id & 0xffffffff)
}

// NewAtom creates a new atom.
func NewAtom(predicate Symbol, terms []Term) *Atom {
	return &Atom{
		Predicate: predicate,
		Terms:     terms,
	}
}

// ID returns the atom ID.
func (a *Atom) ID() AtomID {
	return AtomID((uint64(a.Predicate) << 32) | uint64(len(a.Terms)))
}

func (a *Atom) String() string {
	if a.Predicate.IsExpr() {
		return a.Terms[0].String()
	}

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

// Equals tests if the atoms are equal.
func (a *Atom) Equals(o *Atom) bool {
	return a.EqualsWithMapping(o, make(map[Symbol]Symbol))
}

// EqualsWithMapping tests if the atoms are equal. The mappings are
// updated during the operation.
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

// Rename renames the atom with the env bindings.
func (a *Atom) Rename(env *Bindings) {
	for _, term := range a.Terms {
		term.Rename(env)
	}
}

// Unify unifies the atoms with the env bindings. The function returns
// true if the atoms can be unified and false otherwise.
func (a *Atom) Unify(o *Atom, env *Bindings) bool {
	if a.Predicate != o.Predicate {
		return false
	}
	if len(a.Terms) != len(o.Terms) {
		return false
	}

	for i, t := range a.Terms {
		at := env.Map(t)
		ot := env.Map(o.Terms[i])

		if at.Equals(ot) {
			continue
		}

		if !at.Unify(ot, env) {
			return false
		}
	}

	return true
}

// Clone creates a new independent copy of the atom.
func (a *Atom) Clone() *Atom {
	n := &Atom{
		Predicate: a.Predicate,
		Terms:     make([]Term, len(a.Terms)),
		Flags:     a.Flags,
	}
	for i, term := range a.Terms {
		n.Terms[i] = term.Clone()
	}
	return n
}

// Substitute applies the bindings to the atom in-place and returns
// the modified atom.
func (a *Atom) Substitute(env *Bindings) *Atom {
	for i, term := range a.Terms {
		a.Terms[i] = term.Substitute(env)
	}
	return a
}
