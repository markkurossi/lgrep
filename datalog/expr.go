//
// expr.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"errors"
	"fmt"
	"strconv"
)

var (
	termTrue          = NewTermConstant("true", true)
	errorFalse        = errors.New("false")
	errorDivideByZero = errors.New("divide by zero")
)

type Expr struct {
	Type  ExprType
	Left  *Expr
	Right *Expr
	Value Term
}

type ExprType int

const (
	ExprConstant ExprType = iota
	ExprVariable
	ExprEQ
	ExprGE
	ExprGT
	ExprLE
	ExprLT
	ExprMul
	ExprDiv
	ExprPlus
	ExprMinus
)

var exprTypeNames = map[ExprType]string{
	ExprConstant: "constant",
	ExprVariable: "variable",
	ExprEQ:       "=",
	ExprGE:       ">=",
	ExprGT:       ">",
	ExprLE:       "<=",
	ExprLT:       "<",
	ExprMul:      "*",
	ExprDiv:      "/",
	ExprPlus:     "+",
	ExprMinus:    "-",
}

func (t ExprType) String() string {
	name, ok := exprTypeNames[t]
	if ok {
		return name
	}
	return fmt.Sprintf("{ExprType %d}", t)
}

func (e *Expr) Rename(env *Bindings) {
	switch e.Type {
	case ExprConstant:
	case ExprVariable:
		e.Value.Rename(env)
	default:
		e.Left.Rename(env)
		e.Right.Rename(env)
	}
}

func (e *Expr) Substitute(env *Bindings) {
	switch e.Type {
	case ExprConstant, ExprVariable:
		e.Value = e.Value.Substitute(env)
	default:
		e.Left.Substitute(env)
		e.Right.Substitute(env)
	}
}

func (e *Expr) Eval(env *Bindings) (Term, error) {
	switch e.Type {
	case ExprConstant, ExprVariable:
		return env.Map(e.Value), nil
	}

	left, err := e.Left.Eval(env)
	if err != nil {
		return nil, err
	}
	right, err := e.Right.Eval(env)
	if err != nil {
		return nil, err
	}

	switch e.Type {
	case ExprEQ:
		if left.Unify(right, env) {
			return termTrue, nil
		} else {
			return nil, errorFalse
		}
	}

	leftInt, err := strconv.ParseInt(left.String(), 10, 64)
	if err != nil {
		return nil, err
	}
	rightInt, err := strconv.ParseInt(right.String(), 10, 64)
	if err != nil {
		return nil, err
	}

	switch e.Type {
	case ExprGE:
		if leftInt >= rightInt {
			return termTrue, nil
		} else {
			return nil, errorFalse
		}

	case ExprGT:
		if leftInt > rightInt {
			return termTrue, nil
		} else {
			return nil, errorFalse
		}

	case ExprLE:
		if leftInt <= rightInt {
			return termTrue, nil
		} else {
			return nil, errorFalse
		}

	case ExprLT:
		if leftInt < rightInt {
			return termTrue, nil
		} else {
			return nil, errorFalse
		}

	case ExprMul:
		val := leftInt * rightInt
		return NewTermConstant(fmt.Sprintf("%d", val), false), nil

	case ExprDiv:
		if rightInt == 0 {
			return nil, errorDivideByZero
		}
		val := leftInt / rightInt
		return NewTermConstant(fmt.Sprintf("%d", val), false), nil

	case ExprPlus:
		val := leftInt + rightInt
		return NewTermConstant(fmt.Sprintf("%d", val), false), nil

	case ExprMinus:
		val := leftInt - rightInt
		return NewTermConstant(fmt.Sprintf("%d", val), false), nil
	}

	return nil, fmt.Errorf("Invalid expression '%s'", e)
}

func (e *Expr) Equals(o *Expr) bool {
	if e.Type != o.Type {
		return false
	}
	switch e.Type {
	case ExprConstant, ExprVariable:
		return e.Value.Equals(o.Value)

	default:
		return e.Left.Equals(o.Left) && e.Right.Equals(o.Right)
	}
}

func (e *Expr) Clone() *Expr {
	switch e.Type {
	case ExprConstant, ExprVariable:
		return &Expr{
			Type:  e.Type,
			Value: e.Value,
		}

	default:
		return &Expr{
			Type:  e.Type,
			Left:  e.Left.Clone(),
			Right: e.Right.Clone(),
		}
	}
}

func (e *Expr) String() string {
	switch e.Type {
	case ExprConstant, ExprVariable:
		return e.Value.String()

	default:
		return fmt.Sprintf("%s %s %s",
			e.Left.String(), e.Type.String(), e.Right.String())
	}
}
