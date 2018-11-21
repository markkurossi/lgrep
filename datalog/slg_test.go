//
// slg_test.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"fmt"
	"io"
	"strings"
	"testing"
)

var clauseData = []string{
	"data(a,X). data(a,Y).",
	"data(a,X). data(a,X).",
	"data(a,b). data(a,b).",
	"data(X,Y):-x(X,b). data(Y,Z):-x(Y,b).",
	"data(X,Y):-x(X,b),y(b,Y). data(Y,Z):-x(Y,b),y(b,Z).",
}

func TestClause(t *testing.T) {
	for _, data := range clauseData {
		in := strings.NewReader(data)
		parser := NewParser("data", in)

		c1, _, err := parser.Parse()
		if err != nil {
			t.Fatalf("Clause1: %s", err)
		}

		c2, _, err := parser.Parse()
		if err != nil {
			t.Fatalf("Clause2: %s", err)
		}
		if !c1.Equals(c2) {
			t.Errorf("%s != %s\n", c1, c2)
		}
	}
}

func TestSLG(t *testing.T) {
	in := strings.NewReader(input)
	parser := NewParser("data", in)

	for {
		clause, clauseType, err := parser.Parse()
		if err != nil {
			if err != io.EOF {
				t.Errorf("Parser error: %s", err)
			}
			break
		}
		if false {
			fmt.Printf("%s%s\n", clause, clauseType)
		}

		switch clauseType {
		case ClauseFact:
			DBAdd(clause)

		case ClauseQuery:
			fmt.Printf("%s%s\n", clause, clauseType)
			result := Query(clause.Head)
			for _, r := range result {
				fmt.Printf("=> %s\n", r)
			}
		}
	}
}
