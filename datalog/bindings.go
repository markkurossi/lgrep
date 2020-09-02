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

// Bindings implement symbol bindings.
type Bindings struct {
	arr []binding
}

type binding struct {
	sym Symbol
	val Term
}

// NewBindings creates a new bindings instance.
func NewBindings() *Bindings {
	return &Bindings{}
}

// Size returns the number of bindings.
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

// Clone creates a new independent copy of the bindings.
func (env *Bindings) Clone() *Bindings {
	n := NewBindings()
	for _, b := range env.arr {
		n.arr = append(n.arr, b)
	}
	return n
}

// Map maps the argument term to its current binding in the
// environment. The function returns the mapped value or the argument
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

// Contains tests if the symbols has a binding.
func (env *Bindings) Contains(s Symbol) bool {
	for _, b := range env.arr {
		if b.sym == s {
			return true
		}
	}
	return false
}

// Bind binds the symbol for the term. The function returns true if
// the binding was added and false if the symbol was already bound.
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
