//
// atom.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

type Flags int

const (
	FlagPersistent Flags = 1 << iota
)

type Atom struct {
	Predicate Symbol
	Terms     []Term
	Flags     Flags
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

func (a *Atom) Rename(env Bindings) {
	for _, term := range a.Terms {
		term.Rename(env)
	}
}

func (a *Atom) Unify(o *Atom, env Bindings) *Atom {
	if a.Predicate != o.Predicate {
		return nil
	}
	if len(a.Terms) != len(o.Terms) {
		return nil
	}
	var newTerms []Term

	baseEnv := env.Clone()

	for i, t := range a.Terms {
		at := baseEnv.Map(t)
		ot := baseEnv.Map(o.Terms[i])

		unified := at.Unify(ot, env)
		if unified == nil {
			return nil
		}
		newTerms = append(newTerms, unified)
	}

	return &Atom{
		Predicate: a.Predicate,
		Terms:     newTerms,
		// XXX flags
	}
}

func (a *Atom) Clone() *Atom {
	n := &Atom{
		Predicate: a.Predicate,
		Terms:     make([]Term, len(a.Terms)),
	}
	for i, term := range a.Terms {
		n.Terms[i] = term
	}
	return n
}

// Substitute applies the bindings to the atom in-place and returns
// the modified atom.
func (a *Atom) Substitute(env Bindings) *Atom {
	for i, term := range a.Terms {
		a.Terms[i] = env.Map(term)
	}
	return a
}
