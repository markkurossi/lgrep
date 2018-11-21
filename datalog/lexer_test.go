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
%parent(X, Y)?
%parent(bill, X)?

% rules
ancestor(X, Y) :- parent(X, Y).
ancestor(X, Y) :- parent(X, Z), ancestor(Z, Y).
ancestor(X, john)?

% generate problem of size 10
reachable(X,Y) :- edge(X,Y).
reachable(X,Y) :- edge(X,Z), reachable(Z,Y).
same_clique(X,Y) :- reachable(X,Y), reachable(Y,X).
edge(0, 1).
edge(1, 2).
edge(2, 3).
edge(3, 4).
edge(4, 5).
edge(5, 0).
edge(5, 6).
edge(6, 7).
edge(7, 8).
edge(8, 9).
edge(9, 10).
edge(10, 0).
same_clique(0,10)?
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
