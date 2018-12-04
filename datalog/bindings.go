//
// bindings.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"fmt"
)

type Bindings struct {
	arr []binding
}

type binding struct {
	sym Symbol
	val Term
}

func NewBindings() *Bindings {
	return &Bindings{}
}

func (env *Bindings) Size() int {
	return len(env.arr)
}

func (env *Bindings) String() string {
	var str string

	for _, b := range env.arr {
		if len(str) > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%s->%s", b.sym, b.val)
	}
	return "{" + str + "}"
}

func (env *Bindings) Clone() *Bindings {
	n := NewBindings()
	for _, b := range env.arr {
		n.arr = append(n.arr, b)
	}
	return n
}

// Map maps the argument term to its current binding in the
// environment. The function returns the mapped value of the argument
// term if the environment does not have a mapping for the term.
func (env *Bindings) Map(term Term) Term {
	sym := term.Variable()
	if sym != SymNil {
		for _, b := range env.arr {
			if b.sym == sym {
				return b.val
			}
		}
	}
	return term
}

func (env *Bindings) Contains(s Symbol) bool {
	for _, b := range env.arr {
		if b.sym == s {
			return true
		}
	}
	return false
}

func (env *Bindings) Bind(s Symbol, term Term) bool {
	if env.Contains(s) {
		return false
	}
	env.arr = append(env.arr, binding{
		sym: s,
		val: term,
	})
	return true
}
