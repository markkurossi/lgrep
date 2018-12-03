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
	Unify(t Term, env *Bindings) Term
	Equals(t Term) bool
	String() string
	RenameSLG(env EnvironmentSLG)
	SubstituteSLG(env EnvironmentSLG) Term
	UnifySLG(t Term, env EnvironmentSLG) EnvironmentSLG
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

func (t *TermVariable) RenameSLG(env EnvironmentSLG) {
	env[t.Symbol] = NewTermVariable(newUniqueSymbol())
}

func (t *TermVariable) Rename(env *Bindings) {
	if !env.Contains(t.Symbol) {
		env.Bind(t.Symbol, NewTermVariable(newUniqueSymbol()))
	}
}

func (t *TermVariable) SubstituteSLG(env EnvironmentSLG) Term {
	subst := env[t.Symbol]
	if subst != nil {
		return subst
	}
	return t
}

func (t *TermVariable) UnifySLG(other Term, env EnvironmentSLG) EnvironmentSLG {
	switch o := other.(type) {
	case *TermVariable:
		env[o.Symbol] = t

	case *TermConstant:
		env[t.Symbol] = o
	}
	return env
}

func (t *TermVariable) Unify(other Term, env *Bindings) Term {
	switch o := other.(type) {
	case *TermVariable:
		if t.Symbol == o.Symbol {
			// Same variable.
			return t
		}
		// Unify(T, O): replace O with T
		newT := env.Map(t)
		if !env.Bind(o.Symbol, newT) {
			// O already bound, bind T to O's new binding.
			newO := env.Map(o)
			if !env.Bind(t.Symbol, newO) {
				return nil
			}
			return newO
		}
		return newT

	case *TermConstant:
		// Unify(T, o): assign T to o
		if !env.Bind(t.Symbol, o) {
			// T already bound.
			newT := env.Map(t)
			if o.Equals(newT) {
				// Same binding.
				return newT
			}
			return nil
		}
		return o
	}
	return nil
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

func (t *TermConstant) RenameSLG(env EnvironmentSLG) {
}

func (t *TermConstant) Rename(env *Bindings) {
}

func (t *TermConstant) SubstituteSLG(env EnvironmentSLG) Term {
	return t
}

func (t *TermConstant) UnifySLG(other Term, env EnvironmentSLG) EnvironmentSLG {
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

func (t *TermConstant) Unify(other Term, env *Bindings) Term {
	switch o := other.(type) {
	case *TermVariable:
		// Unify(t, O): assign O to t
		if !env.Bind(o.Symbol, t) {
			// O already bound.
			newO := env.Map(o)
			if t.Equals(newO) {
				// Same binding.
				return newO
			}
			return nil
		}
		return t

	case *TermConstant:
		// Unify(t, o): t must be o
		if t.Value == o.Value {
			return t
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
