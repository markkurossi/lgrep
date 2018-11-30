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

type Bindings map[Symbol]Term

func NewBindings() Bindings {
	return make(Bindings)
}

func (env Bindings) Size() int {
	return len(env)
}

func (env Bindings) String() string {
	var str string

	for k, v := range env {
		if len(str) > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%s->%s", k, v)
	}
	return "{" + str + "}"
}

func (e Bindings) Clone() Bindings {
	n := NewBindings()
	for k, v := range e {
		n[k] = v
	}
	return n
}

// Map maps the argument term to its current binding in the
// environment. The function returns the mapped value of the argument
// term if the environment does not have a mapping for the term.
func (e Bindings) Map(term Term) Term {
	mapped, ok := e[term.Variable()]
	if ok {
		return mapped
	}
	return term
}

func (e Bindings) Contains(s Symbol) bool {
	_, ok := e[s]
	return ok
}

func (e Bindings) Bind(s Symbol, term Term) bool {
	_, ok := e[s]
	if ok {
		return false
	}
	e[s] = term
	return true
}
