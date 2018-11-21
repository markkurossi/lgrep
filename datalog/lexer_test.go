//
// lexer_test.go
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

var input string = `
% Facts
parent(bill, mary).
parent(mary, john).
"="(value,42).
fact.
fact~

% Queries
parent(X, Y)?
parent(bill, X)?

% rules
ancestor(X, Y) :- parent(X, Y).
ancestor(X, Y) :- parent(X, Z), ancestor(Z, Y).
ancestor(X, john)?
`

func TestLexer(t *testing.T) {
	in := strings.NewReader(input)
	lexer := NewLexer("data", in)
	for {
		token, err := lexer.GetToken()
		if err != nil {
			if err != io.EOF {
				t.Errorf("Lexer error: %s", err)
			}
			break
		}
		if false {
			fmt.Printf("Token: %s\n", token)
		}
	}
}
