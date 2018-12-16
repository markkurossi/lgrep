//
// parser_test.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"fmt"
	"strings"
	"testing"
)

var clauseData = []string{
	"data(a,X). data(a,Y).",
	"data(a,X). data(a,X).",
	"data(a,b). data(a,b).",
	"data(X,Y):-x(X,b). data(Y,Z):-x(Y,b).",
	"data(X,Y):-x(X,b),y(b,Y). data(Y,Z):-x(Y,b),y(b,Z).",
	"a(A,B) :- A=B. a(A,B) :- A=B.",
}

func TestClause(t *testing.T) {
	for _, data := range clauseData {
		in := strings.NewReader(data)
		parser := NewParser("data", in)

		c1, _, err := parser.Parse()
		if err != nil {
			t.Fatalf("Clause1: %s\nInput: %s\n", err, data)
		}

		c2, _, err := parser.Parse()
		if err != nil {
			t.Fatalf("Clause2: %s", err)
		}
		if !c1.Equals(c2) {
			t.Errorf("%s != %s\n", c1, c2)
		} else if false {
			fmt.Printf("%s = %s\n", c1, c2)
		}
	}
}
