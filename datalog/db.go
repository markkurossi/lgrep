//
// db.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

// DB defines the clause storage interface.
type DB interface {
	Add(clause *Clause)
	Get(atom *Atom, limits Predicates) []*Clause
	Sync()
}

// MemDB implements the memory DB storage.
type MemDB struct {
	clauses map[AtomID][]*Clause
}

// NewMemDB implements the DB interface with memory storage.
func NewMemDB() DB {
	return &MemDB{
		clauses: make(map[AtomID][]*Clause),
	}
}

// Add implements the DB.Add.
func (db *MemDB) Add(clause *Clause) {
	arr, ok := db.clauses[clause.Head.ID()]
	if !ok {
		arr = make([]*Clause, 0, 10)
	}
	arr = append(arr, clause)
	db.clauses[clause.Head.ID()] = arr
}

// Get implements the DB.Get.
func (db *MemDB) Get(atom *Atom, limits Predicates) []*Clause {
	var result []*Clause
	for _, c := range db.clauses[atom.ID()] {
		if !c.IsFact() || c.Timestamp > limits[atom.ID()] {
			result = append(result, c)
		}
	}
	return result
}

// Sync implements the DB.Sync.
func (db *MemDB) Sync() {
}
