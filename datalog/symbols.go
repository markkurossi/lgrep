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

type Symbol uint32

type SymbolName struct {
	Name       string
	Stringlike bool
}

var (
	m                    = &sync.Mutex{}
	nextSymbolID  Symbol = SymFirstIntern
	symbolsByID          = make(map[Symbol]SymbolName)
	symbolsByName        = make(map[string]Symbol)
)

const (
	SymNil Symbol = iota
	SymEQ
	SymGE
	SymGT
	SymLE
	SymLT
	SymFirstIntern
)

func (s Symbol) IsExpr() bool {
	return s != SymNil && s < SymFirstIntern
}

func (s Symbol) String() string {
	if s.IsExpr() {
		switch s {
		case SymEQ:
			return "="
		case SymGE:
			return ">="
		case SymGT:
			return ">"
		case SymLE:
			return "<="
		case SymLT:
			return "<"
		default:
			return fmt.Sprintf("{expr %d}", s)
		}
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
