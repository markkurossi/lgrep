//
// db.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

var clauses = make(map[Symbol][]*Clause)

func DBAdd(clause *Clause) {
	arr, ok := clauses[clause.Head.Predicate]
	if !ok {
		arr = make([]*Clause, 0, 10)
	}
	arr = append(arr, clause)
	clauses[clause.Head.Predicate] = arr
}

func DBClauses(predicate Symbol) []*Clause {
	return clauses[predicate]
}
