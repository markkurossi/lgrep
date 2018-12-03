//
// atom_test.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"testing"
)

func symbol(name string) Symbol {
	sym, _ := Intern(name, false)
	return sym
}

func atom1() *Atom {
	terms := []Term{
		NewTermConstant("a", false),
		NewTermConstant("b", false),
		NewTermConstant("c", false),
	}
	return NewAtom(symbol("np"), terms)
}

func atom2() *Atom {
	terms := []Term{
		NewTermVariable(symbol("A")),
		NewTermVariable(symbol("B")),
		NewTermVariable(symbol("C")),
	}
	return NewAtom(symbol("np"), terms)
}

func BenchmarkUnify(b *testing.B) {
	a1 := atom1()
	a2 := atom2()

	for i := 0; i < b.N; i++ {
		env := NewBindings()
		if !a1.Unify(a2, env) {
			b.Fatalf("%s.Unify(%s) failed\n", a1, a2)
		}
		unified := a1.Clone().Substitute(env)
		if unified == nil {
			b.Fatalf("%s.Substitute(%s) failed\n", a1, env)
		}
	}
}

func BenchmarkUnifySLG(b *testing.B) {
	a1 := atom1()
	a2 := atom2()

	for i := 0; i < b.N; i++ {
		env := a1.UnifySLG(a2)
		if env == nil {
			b.Fatalf("%s.UnifySLG(%s) failed\n", a1, a2)
		}
		unified := a1.SubstituteSLG(env)
		if unified == nil {
			b.Fatalf("%s.SubstituteSLG(%s) failed\n", a1, env)
		}
	}
}
