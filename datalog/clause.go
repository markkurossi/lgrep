//
// clause.go
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
		Timestamp: time.Now().UnixNano(),
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

type Predicates map[Symbol]int64

func (p Predicates) String() string {
	var str string
	for k, v := range p {
		if len(str) > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%s=%d", k, v)
	}
	return "[" + str + "]"
}

// Predicates returns all predicates, used or linked by this
// clause. The predicates are returned in a map from the predicate
// symbol to int64 so that the returned map can be used, for example,
// to implement database search limiting.
func (c *Clause) Predicates(db DB, flags Flags) Predicates {
	result := make(Predicates)
	pending := []*Clause{c}

	for len(pending) > 0 {
		var newPending []*Clause
		for _, c := range pending {
			atoms := []*Atom{c.Head}
			atoms = append(atoms, c.Body...)
			for _, atom := range atoms {
				if atom.Flags != flags {
					continue
				}
				_, ok := result[atom.Predicate]
				if !ok {
					result[atom.Predicate] = 0
					cls := db.Get(atom.Predicate, nil)
					if len(cls) > 0 {
						newPending = append(newPending, cls[0])
					}
				}
			}
		}

		pending = newPending
	}

	return result
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
