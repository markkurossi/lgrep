//
// query_test.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
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
	test{
		file:   "clique100.pl",
		result: "same_clique(0, 100).",
	},
	test{
		file:   "graph100.pl",
		result: "reachable(11, 99).",
	},
	test{
		file:   "induction100.pl",
		result: "q(100).",
	},
	test{
		file: "ship.pl",
		result: `
ship_to(tea, london).
ship_to(bread, paris).
ship_to(flowers, "San Francisco").
ship_to(sausage, munich).
ship_to(horse, seoul).
`,
	},
	test{
		file: "small.pl",
		result: `
ancestor(brad, john).
ancestor(brad, ann).
ancestor(brad, bill).
`,
	},
	test{
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
	test{
		file: "laps.dl",
		result: `
permit(rams, store, rams_couch).
permit(will, fetch, rams_couch).
`,
	},
	test{
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
	test{
		file:   "pq.dl",
		result: `q(a).`,
	},
	test{
		file:   "says.dl",
		result: `says(tpme1, m1).`,
	},
	test{
		file: "tc.dl",
		result: `
r(a, b).
r(a, c).
`,
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
				if false {
					result = QuerySLG(clause.Head, db, nil)
				} else {
					result = Execute(clause.Head, db, nil)
				}
				expected := parseString(t, test.result)
				if len(result) != len(expected) {
					t.Errorf("Unexpected number of results: %d vs. %d\n",
						len(result), len(expected))
				}
				if !result.Equals(expected) {
					t.Errorf("Unexpected result %s vs. %s\n", result, expected)
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
