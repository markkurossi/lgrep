//
// unify_test.go
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

type UnifyTest struct {
	A string
	B string
	R string
	E map[string]string
}

var unifyTests = []UnifyTest{
	{
		A: "p1(X).",
		B: "p2(Y).",
		R: "",
	},
	{
		A: "p(X,Y).",
		B: "p(Y).",
		R: "",
	},
	{
		A: "p(X).",
		B: "p(a).",
		R: "p(a).",
		E: map[string]string{
			"X": "a",
		},
	},
	{
		A: "p(a).",
		B: "p(X).",
		R: "p(a).",
		E: map[string]string{
			"X": "a",
		},
	},
	{
		A: "p(a).",
		B: "p(b).",
		R: "",
	},
	{
		A: "p(X).",
		B: "p(Y).",
		R: "p(X).",
		E: map[string]string{
			"Y": "X",
		},
	},
	{
		A: "p(X,Y).",
		B: "p(Q,W).",
		R: "p(X,Y).",
		E: map[string]string{
			"Q": "X",
			"W": "Y",
		},
	},
	{
		A: "p(a,Y).",
		B: "p(Q,W).",
		R: "p(a,Y).",
		E: map[string]string{
			"Q": "a",
			"W": "Y",
		},
	},
	{
		A: "p(Q,W).",
		B: "p(a,Y).",
		R: "p(a,W).",
		E: map[string]string{
			"Q": "a",
			"Y": "W",
		},
	},
	// 9
	{
		A: "p(X,Y).",
		B: "p(Q,Q).",
		R: "p(Y,Y).",
		E: map[string]string{
			"Q": "X",
			"X": "Y",
		},
	},
	{
		A: "p(a,Y,X).",
		B: "p(Q,z,Q).",
		R: "p(a,z,a).",
		E: map[string]string{
			"Q": "a",
			"Y": "z",
			"X": "a",
		},
	},
	{
		A: "p(Q,z,Q).",
		B: "p(a,Y,X).",
		R: "p(a,z,a).",
		E: map[string]string{
			"Q": "a",
			"Y": "z",
			"X": "a",
		},
	},
	{
		A: "p(X,a).",
		B: "p(a,X).",
		R: "p(a,a).",
		E: map[string]string{
			"X": "a",
		},
	},
	{
		A: "p(X,X).",
		B: "p(a,a).",
		R: "p(a,a).",
		E: map[string]string{
			"X": "a",
		},
	},
	{
		A: "p(a,a).",
		B: "p(X,X).",
		R: "p(a,a).",
		E: map[string]string{
			"X": "a",
		},
	},
	{
		A: "p(X,Y,Z).",
		B: "p(a,X,Y).",
		R: "p(a,a,a).",
		E: map[string]string{
			"X": "a",
			"Y": "a",
			"Z": "a",
		},
	},
	{
		A: "p(a,X,Y).",
		B: "p(X,Y,Z).",
		R: "p(a,a,a).",
		E: map[string]string{
			"X": "a",
			"Y": "a",
			"Z": "a",
		},
	},
}

func TestUnify(t *testing.T) {
	for index, test := range unifyTests {
		if index != 9 && false {
			continue
		}
		a, _, err := parseClause(test.A)
		if err != nil {
			t.Fatalf("Failed to parse clause '%s': %s\n", test.A, err)
		}
		b, _, err := parseClause(test.B)
		if err != nil {
			t.Fatalf("Failed to parse clause '%s': %s\n", test.B, err)
		}
		env := NewBindings()
		if !a.Head.Unify(b.Head, env) {
			if len(test.R) > 0 {
				t.Errorf("Unify(%s, %s) failed, expected %s\n", a, b, test.R)
			}
		} else {
			if len(test.R) == 0 {
				t.Errorf("Unify(%s, %s) should have failed\n", a, b)
				continue
			}

			r, _, err := parseClause(test.R)
			if err != nil {
				t.Fatalf("Failed to parser clause '%s': %s\n", test.R, err)
			}
			unified := a.Head.Clone().Substitute(env)
			if !unified.Equals(r.Head) {
				t.Errorf("Unified %s not equal to expected %s\n",
					unified, r.Head)
			}

			if false {
				fmt.Printf("Unify(%s,%s) => %s %s\n",
					test.A, test.B, unified, env)
			}

			// Check env bindings.
			for k, v := range test.E {
				sym, _ := Intern(k, false)
				term := env.Map(NewTermVariable(sym))
				if term == nil {
					t.Errorf("%v: Symbol %s has no binding in environment %s\n",
						test, k, env)
				} else {
					if v != term.String() {
						t.Errorf(
							"%d: Symbol %s: invalid binding in %s: %s vs %s\n",
							index, k, env, v, term)
					}
				}
			}
		}
	}
}

func parseClause(input string) (*Clause, ClauseType, error) {
	return NewParser("{data}", strings.NewReader(input)).Parse()
}
