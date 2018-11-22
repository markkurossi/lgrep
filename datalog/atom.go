//
// atom.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

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
