//
// term.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

type TermType int

const (
	Variable TermType = iota
	Constant
)

type Term interface {
	Type() TermType
	Variable() Symbol
	Rename(env *Bindings)
	Unify(t Term, env *Bindings) bool
	Equals(t Term) bool
	String() string
}

type TermVariable struct {
	Symbol Symbol
}

func NewTermVariable(symbol Symbol) Term {
	return &TermVariable{
		Symbol: symbol,
	}
}

func (t *TermVariable) Type() TermType {
	return Variable
}

func (t *TermVariable) Variable() Symbol {
	return t.Symbol
}

func (t *TermVariable) Rename(env *Bindings) {
	if !env.Contains(t.Symbol) {
		env.Bind(t.Symbol, NewTermVariable(newUniqueSymbol()))
	}
}

func (t *TermVariable) Unify(other Term, env *Bindings) bool {
	switch o := other.(type) {
	case *TermVariable:
		if t.Symbol == o.Symbol {
			// Same variable.
			return true
		}
		// Unify(T, O): replace O with T
		newT := env.Map(t)
		if env.Bind(o.Symbol, newT) {
			return true
		}
		// O already bound, bind T to O's new binding.
		newO := env.Map(o)
		if env.Bind(t.Symbol, newO) {
			return true
		}
		return false

	case *TermConstant:
		// Unify(T, o): assign T to o
		if env.Bind(t.Symbol, o) {
			return true
		}
		// T already bound.
		newT := env.Map(t)
		if o.Equals(newT) {
			// Same binding.
			return true
		}
		return false
	}
	return false
}

func (t *TermVariable) Equals(other Term) bool {
	switch o := other.(type) {
	case *TermVariable:
		return t.Symbol == o.Symbol
	}
	return false
}

func (t *TermVariable) String() string {
	return t.Symbol.String()
}

type TermConstant struct {
	Value      string
	Stringlike bool
}

func NewTermConstant(value string, stringlike bool) Term {
	return &TermConstant{
		Value:      value,
		Stringlike: stringlike,
	}
}

func (t *TermConstant) Type() TermType {
	return Constant
}

func (t *TermConstant) Variable() Symbol {
	return SymNil
}

func (t *TermConstant) Rename(env *Bindings) {
}

func (t *TermConstant) Unify(other Term, env *Bindings) bool {
	switch o := other.(type) {
	case *TermVariable:
		// Unify(t, O): assign O to t
		if env.Bind(o.Symbol, t) {
			return true
		}
		// O already bound.
		newO := env.Map(o)
		if t.Equals(newO) {
			// Same binding.
			return true
		}
		return false

	case *TermConstant:
		// Unify(t, o): t must be o
		if t.Value == o.Value {
			return true
		}
	}
	return false
}

func (t *TermConstant) Equals(other Term) bool {
	switch o := other.(type) {
	case *TermConstant:
		return t.Value == o.Value
	}

	return false
}

func (t *TermConstant) String() string {
	if t.Stringlike {
		return Stringify(t.Value)
	} else {
		return t.Value
	}
}
