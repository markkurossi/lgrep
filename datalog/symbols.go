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
	nextSymbolID  Symbol = 1
	symbolsByID          = make(map[Symbol]SymbolName)
	symbolsByName        = make(map[string]Symbol)
)

func (s Symbol) String() string {
	m.Lock()
	name, ok := symbolsByID[s]
	m.Unlock()
	if ok {
		if name.Stringlike {
			return stringify(name.Name)
		}
		return name.Name
	}
	// Unique variable.
	return fmt.Sprintf("{%d}", s)
}

func intern(symbol string, stringlike bool) Symbol {
	m.Lock()
	id, ok := symbolsByName[symbol]
	if !ok {
		id = nextSymbolID
		nextSymbolID++
		symbolsByName[symbol] = id
		symbolsByID[id] = SymbolName{
			Name:       symbol,
			Stringlike: stringlike,
		}
	}
	m.Unlock()
	return id
}

func newUniqueSymbol() Symbol {
	m.Lock()
	symbol := nextSymbolID
	nextSymbolID++
	m.Unlock()
	return symbol
}
