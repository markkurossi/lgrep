//
// environment.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"fmt"
)

type Environment map[Symbol]Term

func NewEnvironment() Environment {
	return make(Environment)
}

func (env Environment) String() string {
	var str string

	for k, v := range env {
		if len(str) > 0 {
			str += ", "
		}
		str += fmt.Sprintf("%s=%s", k, v)
	}
	return "[" + str + "]"
}

func (e Environment) Clone() Environment {
	n := NewEnvironment()
	for k, v := range e {
		n[k] = v
	}
	return n
}

// Map maps the argument term to its current binding in the
// environment. The function returns the mapped value of the argument
// term if the environment does not have a mapping for the term.
func (e Environment) Map(term Term) Term {
	mapped, ok := e[term.Variable()]
	if ok {
		return mapped
	}
	return term
}

func (e Environment) Bind(s Symbol, term Term) bool {
	_, ok := e[s]
	if ok {
		return false
	}
	e[s] = term
	return true
}
