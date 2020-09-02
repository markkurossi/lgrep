//
// symbols.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"fmt"
	"sync"
)

// Symbol specifies symbol instances.
type Symbol uint32

// SymbolName implements symbol names.
type SymbolName struct {
	Name       string
	Stringlike bool
}

var (
	m             = &sync.Mutex{}
	nextSymbolID  = SymFirstIntern
	symbolsByID   = make(map[Symbol]SymbolName)
	symbolsByName = make(map[string]Symbol)
)

// Symbol types.
const (
	SymNil Symbol = iota
	SymExpr
	SymFirstIntern
)

// IsExpr tests if the symbol is an expression.
func (s Symbol) IsExpr() bool {
	return s == SymExpr
}

func (s Symbol) String() string {
	switch s {
	case SymNil:
		return "{nil}"
	case SymExpr:
		return "{expression}"
	}

	m.Lock()
	name, ok := symbolsByID[s]
	m.Unlock()
	if ok {
		if name.Stringlike {
			return Stringify(name.Name)
		}
		return name.Name
	}
	// Unique variable.
	return fmt.Sprintf(":%d", s)
}

// Intern interns the argument string so that all equal strings return
// the same Symbol instance.
func Intern(value string, stringlike bool) (Symbol, string) {
	m.Lock()
	var name SymbolName
	id, ok := symbolsByName[value]
	if ok {
		name = symbolsByID[id]
	} else {
		id = nextSymbolID
		nextSymbolID++
		symbolsByName[value] = id

		name = SymbolName{
			Name:       value,
			Stringlike: stringlike,
		}
		symbolsByID[id] = name
	}
	m.Unlock()
	return id, name.Name
}

func newUniqueSymbol() Symbol {
	m.Lock()
	symbol := nextSymbolID
	nextSymbolID++
	m.Unlock()
	return symbol
}
