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
	"strconv"
)

type Flags int

const (
	FlagPersistent Flags = 1 << iota
)

type Atom struct {
	Predicate Symbol
	Terms     []Term
	Flags     Flags
}

type AtomID uint64

func (id AtomID) String() string {
	return fmt.Sprintf("%s/%d", id.Symbol(), id.Arity())
}

func (id AtomID) Symbol() Symbol {
	return Symbol(id >> 32)
}

func (id AtomID) Arity() int {
	return int(id & 0xffffffff)
}

func NewAtom(predicate Symbol, terms []Term) *Atom {
	return &Atom{
		Predicate: predicate,
		Terms:     terms,
	}
}

func (a *Atom) ID() AtomID {
	return AtomID((uint64(a.Predicate) << 32) | uint64(len(a.Terms)))
}

func (a *Atom) String() string {
	if a.Predicate.IsExpr() {
		return fmt.Sprintf("%s %s %s", a.Terms[0], a.Predicate, a.Terms[1])
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

func (a *Atom) Equals(o *Atom) bool {
	return a.EqualsWithMapping(o, make(map[Symbol]Symbol))
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

func (a *Atom) Rename(env *Bindings) {
	for _, term := range a.Terms {
		term.Rename(env)
	}
}

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

func (a *Atom) Clone() *Atom {
	n := &Atom{
		Predicate: a.Predicate,
		Terms:     make([]Term, len(a.Terms)),
		Flags:     a.Flags,
	}
	for i, term := range a.Terms {
		n.Terms[i] = term
	}
	return n
}

// Substitute applies the bindings to the atom in-place and returns
// the modified atom.
func (a *Atom) Substitute(env *Bindings) *Atom {
	for i, term := range a.Terms {
		a.Terms[i] = env.Map(term)
	}
	return a
}

func (a *Atom) Eval(env *Bindings) bool {
	if len(a.Terms) != 2 {
		return false
	}

	switch a.Predicate {
	case SymEQ:
		return a.Terms[0].Unify(a.Terms[1], env)

	case SymGE, SymGT, SymLE, SymLT:
		v1, err := strconv.Atoi(env.Map(a.Terms[0]).String())
		if err != nil {
			fmt.Printf("%s: %s\n", a.Predicate, err)
			return false
		}
		v2, err := strconv.Atoi(env.Map(a.Terms[1]).String())
		if err != nil {
			fmt.Printf(">: %s\n", err)
			return false
		}
		switch a.Predicate {
		case SymGT:
			return v1 > v2
		case SymGE:
			return v1 >= v2
		case SymLE:
			return v1 <= v2
		case SymLT:
			return v1 < v2
		default:
			return false
		}

	default:
		return false
	}
}
