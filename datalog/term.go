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
	Rename(env Environment)
	Substitute(env Environment) Term
	Unify(t Term, env Environment) Environment
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

func (t *TermVariable) Rename(env Environment) {
	env[t.Symbol] = NewTermVariable(newUniqueSymbol())
}

func (t *TermVariable) Substitute(env Environment) Term {
	subst := env[t.Symbol]
	if subst != nil {
		return subst
	}
	return t
}

func (t *TermVariable) Unify(other Term, env Environment) Environment {
	switch o := other.(type) {
	case *TermVariable:
		env[o.Symbol] = t

	case *TermConstant:
		env[t.Symbol] = o
	}
	return env
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
	return NilSymbol
}

func (t *TermConstant) Rename(env Environment) {
}

func (t *TermConstant) Substitute(env Environment) Term {
	return t
}

func (t *TermConstant) Unify(other Term, env Environment) Environment {
	switch o := other.(type) {
	case *TermVariable:
		env[o.Symbol] = t
		return env

	case *TermConstant:
		if t.Value == o.Value {
			return env
		}
	}

	return nil
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
