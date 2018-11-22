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
	Get(predicate Symbol, from int64) []*Clause
}

type MemDB struct {
	clauses map[Symbol][]*Clause
}

func NewMemDB() DB {
	return &MemDB{
		clauses: make(map[Symbol][]*Clause),
	}
}

func (db *MemDB) Add(clause *Clause) {
	arr, ok := db.clauses[clause.Head.Predicate]
	if !ok {
		arr = make([]*Clause, 0, 10)
	}
	arr = append(arr, clause)
	db.clauses[clause.Head.Predicate] = arr
}

func (db *MemDB) Get(predicate Symbol, from int64) []*Clause {
	var result []*Clause
	for _, c := range db.clauses[predicate] {
		if !c.Fact() || c.Timestamp > from {
			result = append(result, c)
		}
	}
	return result
}
