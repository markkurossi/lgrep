//
// query_test.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"fmt"
	"io"
	"os"
	"strings"
	"testing"
)

type test struct {
	file   string
	result string
}

var testFiles = []test{
	{
		file:   "clique100.pl",
		result: "same_clique(0, 100).",
	},
	{
		file:   "graph100.pl",
		result: "reachable(11, 99).",
	},
	{
		file:   "induction100.pl",
		result: "q(100).",
	},
	{
		file: "ship.pl",
		result: `
ship_to(tea, london).
ship_to(bread, paris).
ship_to(flowers, "San Francisco").
ship_to(sausage, munich).
ship_to(horse, seoul).
`,
	},
	{
		file: "small.pl",
		result: `
ancestor(brad, john).
ancestor(brad, ann).
ancestor(brad, bill).
`,
	},
	{
		file: "bidipath.dl",
		result: `
path(a, b).
path(b, c).
path(c, d).
path(d, a).
path(a, c).
path(a, d).
path(a, a).
path(b, d).
path(b, a).
path(b, b).
path(c, a).
path(c, b).
path(c, c).
path(d, b).
path(d, c).
path(d, d).
`,
	},
	{
		file: "laps.dl",
		result: `
permit(rams, store, rams_couch).
permit(will, fetch, rams_couch).
`,
	},
	{
		file: "path.dl",
		result: `
path(a, b).
path(b, c).
path(c, d).
path(d, a).
path(a, c).
path(a, d).
path(a, a).
path(b, d).
path(b, a).
path(b, b).
path(c, a).
path(c, b).
path(c, c).
path(d, b).
path(d, c).
path(d, d).
`,
	},
	{
		file:   "pq.dl",
		result: `q(a).`,
	},
	{
		file:   "says.dl",
		result: `says(tpme1, m1).`,
	},
	{
		file: "tc.dl",
		result: `
r(a, b).
r(a, c).
`,
	},
	{
		file: "ancestor.dl",
		result: `
ancestor(bob, douglas).
ancestor(bob, john).
ancestor(ebbon, bob).
ancestor(ebbon, douglas).
ancestor(ebbon, john).
ancestor(john, douglas).
`,
	},
	{
		file: "selection.dl",
		result: `
ans(alpha, alpha, 1, 7).
ans(beta, beta, 23, 10).
`,
	},
	{
		file:   "expr.dl",
		result: `mismatch(42).`,
	},
	{
		file:   "expr-add.dl",
		result: `add(100,50,150).`,
	},
	{
		file:   "expr-sub.dl",
		result: `sub(100,50,50).`,
	},
	{
		file:   "expr-mul.dl",
		result: `mul(100,50,5000).`,
	},
	{
		file:   "expr-div.dl",
		result: `div(100,50,2).`,
	},
}

func TestData(t *testing.T) {
	for _, test := range testFiles {
		db := NewMemDB()

		filename := "test-data/" + test.file
		f, err := os.Open(filename)
		if err != nil {
			t.Fatalf("Failed to open test file %s: %s\n", filename, err)
		}
		defer f.Close()

		parser := NewParser(filename, f)
		for {
			clause, clauseType, err := parser.Parse()
			if err != nil {
				if err == io.EOF {
					break
				}
				t.Fatalf("%s: %s\n", filename, err)
			}

			switch clauseType {
			case ClauseFact:
				db.Add(clause)

			case ClauseQuery:
				var result ResultSet
				result = Execute(clause.Head, db, nil)
				expected := parseString(t, test.result)
				if len(result) != len(expected) {
					t.Errorf("%s: Unexpected number of results: %d vs. %d\n",
						test.file, len(result), len(expected))
				}
				if !result.Equals(expected) {
					t.Errorf("%s: Unexpected result %s vs. %s\n",
						test.file, result, expected)
				}
			}
		}
	}
}

func parseString(t *testing.T, input string) ResultSet {
	in := strings.NewReader(input)
	parser := NewParser("data", in)
	var result ResultSet
	for {
		clause, clauseType, err := parser.Parse()
		if err != nil {
			if err != io.EOF {
				t.Fatalf("Failed to parse expected results: %s\n", err)
			}
			break
		}
		switch clauseType {
		case ClauseFact:
			result = append(result, clause)
		}
	}
	return result
}

type ResultSet []*Clause

func (rs ResultSet) Contains(clause *Clause) bool {
	for _, c := range rs {
		if c.Equals(clause) {
			return true
		}
	}
	return false
}

func (rs ResultSet) Equals(o ResultSet) bool {
	for _, c := range rs {
		if !o.Contains(c) {
			return false
		}
	}
	for _, c := range o {
		if !rs.Contains(c) {
			return false
		}
	}
	return true
}

func BenchmarkEvalClique1000(b *testing.B) {
	benchmarkEval(b, "test-data/clique1000.pl")
}

func BenchmarkEvalInduction1000(b *testing.B) {
	benchmarkEval(b, "test-data/induction1000.pl")
}

func benchmarkEval(b *testing.B, file string) {
	for i := 0; i < b.N; i++ {
		f, err := os.Open(file)
		if err != nil {
			b.Errorf("Failed to open test file %s: %s", file, err)
			return
		}
		defer f.Close()

		parser := NewParser(file, f)
		db := NewMemDB()
		for {
			clause, clauseType, err := parser.Parse()
			if err != nil {
				if err != io.EOF {
					b.Errorf("Parse failed: %v", err)
					return
				}
				break
			}
			switch clauseType {
			case ClauseFact:
				db.Add(clause)

			case ClauseQuery:
				result := Execute(clause.Head, db, nil)
				if false {
					for _, r := range result {
						fmt.Printf("=> %s\n", r)
					}
				}
			}
		}
	}
}
