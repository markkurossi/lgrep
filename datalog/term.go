//
// term.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

// TermType specifies term types.
type TermType int

// Term types.
const (
	Variable TermType = iota
	Constant
	Expression
)

// Term implements a datalog term.
type Term interface {
	// Type returns the term type.
	Type() TermType
	// Variable returns the term as variable.
	Variable() Symbol
	// Rename renames all environment variables in the term.
	Rename(env *Bindings)
	// Substitute substitutes all instances of environment variables
	// in the term.
	Substitute(env *Bindings) Term
	// Unify unifies terms updating environment.
	Unify(t Term, env *Bindings) bool
	// Equals tests if two termas are equal.
	Equals(t Term) bool
	// Clone creates a copy of the term.
	Clone() Term
	String() string
}

// TermVariable implements variable terms.
type TermVariable struct {
	Symbol Symbol
}

// NewTermVariable creates a new variable term.
func NewTermVariable(symbol Symbol) Term {
	return &TermVariable{
		Symbol: symbol,
	}
}

// Type implements Term.Type.
func (t *TermVariable) Type() TermType {
	return Variable
}

// Variable implements Term.Variable.
func (t *TermVariable) Variable() Symbol {
	return t.Symbol
}

// Rename implements Term.Rename.
func (t *TermVariable) Rename(env *Bindings) {
	if !env.Contains(t.Symbol) {
		env.Bind(t.Symbol, NewTermVariable(newUniqueSymbol()))
	}
}

// Substitute implements Term.Substitute.
func (t *TermVariable) Substitute(env *Bindings) Term {
	return env.Map(t)
}

// Unify implements Term.Unify.
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

// Equals implements Term.Equals.
func (t *TermVariable) Equals(other Term) bool {
	switch o := other.(type) {
	case *TermVariable:
		return t.Symbol == o.Symbol
	}
	return false
}

// Clone implements Term.Clone.
func (t *TermVariable) Clone() Term {
	return t
}

func (t *TermVariable) String() string {
	return t.Symbol.String()
}

// TermConstant implements constant terms.
type TermConstant struct {
	Value      string
	Stringlike bool
}

// NewTermConstant creates a new constant term.
func NewTermConstant(value string, stringlike bool) Term {
	return &TermConstant{
		Value:      value,
		Stringlike: stringlike,
	}
}

// Type implements Term.Type.
func (t *TermConstant) Type() TermType {
	return Constant
}

// Variable implemnets Term.Variable.
func (t *TermConstant) Variable() Symbol {
	return SymNil
}

// Rename implements Term.Rename.
func (t *TermConstant) Rename(env *Bindings) {
}

// Substitute implements Term.Substitute.
func (t *TermConstant) Substitute(env *Bindings) Term {
	return t
}

// Unify implements Term.Unify.
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

// Equals implements Term.Equals.
func (t *TermConstant) Equals(other Term) bool {
	switch o := other.(type) {
	case *TermConstant:
		return t.Value == o.Value
	}
	return false
}

// Clone implements Term.Clone.
func (t *TermConstant) Clone() Term {
	return t
}

func (t *TermConstant) String() string {
	if t.Stringlike {
		return Stringify(t.Value)
	}
	return t.Value
}

// TermExpression implements expression terms.
type TermExpression struct {
	Expr *Expr
}

// NewTermExpression creates a new expression term.
func NewTermExpression(expr *Expr) Term {
	return &TermExpression{
		Expr: expr,
	}
}

// Type implements Term.Type.
func (t *TermExpression) Type() TermType {
	return Expression
}

// Variable implements Term.Variable.
func (t *TermExpression) Variable() Symbol {
	return SymNil
}

// Rename implements Term.Rename.
func (t *TermExpression) Rename(env *Bindings) {
	t.Expr.Rename(env)
}

// Substitute implements Term.Substitute.
func (t *TermExpression) Substitute(env *Bindings) Term {
	t.Expr.Substitute(env)
	return t
}

// Unify implements Term.Unify.
func (t *TermExpression) Unify(other Term, env *Bindings) bool {
	val, err := t.Expr.Eval(env)
	if err != nil {
		return false
	}

	return val.Unify(other, env)
}

// Equals implements Term.Equals.
func (t *TermExpression) Equals(other Term) bool {
	switch o := other.(type) {
	case *TermExpression:
		return t.Expr.Equals(o.Expr)
	}
	return false
}

// Clone implements Term.Clone.
func (t *TermExpression) Clone() Term {
	return NewTermExpression(t.Expr.Clone())
}

func (t *TermExpression) String() string {
	return t.Expr.String()
}
