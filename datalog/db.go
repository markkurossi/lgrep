//
// db.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

type DB interface {
	Add(clause *Clause)
	Get(atom *Atom, limits Predicates) []*Clause
	Sync()
}

type MemDB struct {
	clauses map[AtomID][]*Clause
}

func NewMemDB() DB {
	return &MemDB{
		clauses: make(map[AtomID][]*Clause),
	}
}

func (db *MemDB) Add(clause *Clause) {
	arr, ok := db.clauses[clause.Head.ID()]
	if !ok {
		arr = make([]*Clause, 0, 10)
	}
	arr = append(arr, clause)
	db.clauses[clause.Head.ID()] = arr
}

func (db *MemDB) Get(atom *Atom, limits Predicates) []*Clause {
	var result []*Clause
	for _, c := range db.clauses[atom.ID()] {
		if !c.IsFact() || c.Timestamp > limits[atom.ID()] {
			result = append(result, c)
		}
	}
	return result
}

func (db *MemDB) Sync() {
}
