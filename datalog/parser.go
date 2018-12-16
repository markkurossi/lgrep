//
// parser.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"fmt"
	"io"
)

type Parser struct {
	lexer *Lexer
	ungot *Token
}

func NewParser(inputName string, input io.Reader) *Parser {
	return &Parser{
		lexer: NewLexer(inputName, input),
	}
}

func (p *Parser) getToken() (*Token, error) {
	if p.ungot != nil {
		ret := p.ungot
		p.ungot = nil
		return ret, nil
	}
	return p.lexer.GetToken()
}

func (p *Parser) ungetToken(t *Token) {
	p.ungot = t
}

func (p *Parser) peekToken() (TokenType, error) {
	t, err := p.getToken()
	if err != nil {
		return TokenError, err
	}
	p.ungetToken(t)
	return t.Type, nil
}

func (p *Parser) Parse() (clause *Clause, clauseType ClauseType, err error) {
	var atom *Atom
	atom, err = p.parseAtom()
	if err != nil {
		return
	}
	if atom.Predicate.IsExpr() {
		clauseType = ClauseError
		err = fmt.Errorf("Invalid clause head: %s", atom)
		return
	}
	clause = NewClause(atom, nil)

	var token *Token
	token, err = p.getToken()
	if err != nil {
		return
	}

	if token.Type == TokenArrow {
		// Parse body.
		for {
			atom, err = p.parseAtom()
			if err != nil {
				return
			}
			clause.Body = append(clause.Body, atom)

			var next TokenType
			next, err = p.peekToken()
			if err != nil {
				return
			}
			if next != ',' {
				break
			}
			_, err = p.getToken()
			if err != nil {
				return
			}
		}
		token, err = p.getToken()
		if err != nil {
			return
		}
	}

	switch token.Type {
	case '.':
		return clause, ClauseFact, nil
	case '~':
		return clause, ClauseRetract, nil
	case '?':
		return clause, ClauseQuery, nil
	default:
		return nil, ClauseError, fmt.Errorf("%s: invalid clause type: %s",
			token.Position, token)
	}
}

func (p *Parser) parseAtom() (*Atom, error) {
	token, err := p.getToken()
	if err != nil {
		return nil, err
	}
	next, err := p.peekToken()
	if err != nil {
		return nil, err
	}
	if next.IsExpr() {
		expr, err := p.parseExpr(token)
		if err != nil {
			return nil, err
		}
		return NewAtom(SymExpr, []Term{NewTermExpression(expr)}), nil
	}

	if token.Type != TokenIdentifier && token.Type != TokenString {
		return nil, fmt.Errorf("%s: unexpected token: %s",
			token.Position, token)
	}

	symbol, _ := Intern(token.Value, token.Type == TokenString)
	atom := &Atom{
		Predicate: symbol,
	}

	next, err = p.peekToken()
	if err != nil {
		return nil, err
	}
	if next == '(' {
		// Parse terms
		_, err = p.getToken()
		if err != nil {
			return nil, err
		}
		for {
			token, err = p.getToken()
			if err != nil {
				return nil, err
			}
			var term Term
			switch token.Type {
			case TokenVariable:
				symbol, _ := Intern(token.Value, false)
				term = NewTermVariable(symbol)

			case TokenWildcard:
				term = NewTermVariable(newUniqueSymbol())

			case TokenIdentifier:
				term = NewTermConstant(token.Value, false)

			case TokenString:
				term = NewTermConstant(token.Value, true)

			default:
				return nil, fmt.Errorf("%s: invalid predicate symbol: %s",
					token.Position, token)
			}
			atom.Terms = append(atom.Terms, term)

			token, err = p.getToken()
			if token.Type == ')' {
				break
			}
			if token.Type != ',' {
				return nil, fmt.Errorf("%s: expected ',' or ')': %s",
					token.Position, token)
			}
		}
		next, err := p.peekToken()
		if err != nil {
			return nil, err
		}
		if next == TokenIdentifier {
			token, err = p.getToken()
			if err != nil {
				return nil, err
			}
			for _, r := range []rune(token.Value) {
				switch r {
				case 'p':
					atom.Flags |= FlagPersistent

				default:
					return nil, fmt.Errorf("Invalid flag %c", r)
				}
			}
		}
	}

	return atom, nil
}

func (p *Parser) parseExpr(left *Token) (*Expr, error) {
	expr1, err := p.parseComparison(left)
	if err != nil {
		return nil, err
	}
	next, err := p.peekToken()
	if err != nil {
		return nil, err
	}
	if next != TokenEQ {
		return expr1, nil
	}
	_, err = p.getToken()
	if err != nil {
		return nil, err
	}
	right, err := p.getToken()
	if err != nil {
		return nil, err
	}
	expr2, err := p.parseExpr(right)
	if err != nil {
		return nil, err
	}
	return &Expr{
		Type:  ExprEQ,
		Left:  expr1,
		Right: expr2,
	}, nil
}

func (p *Parser) parseComparison(left *Token) (*Expr, error) {
	expr1, err := p.parseAdditive(left)
	if err != nil {
		return nil, err
	}
	next, err := p.peekToken()
	if err != nil {
		return nil, err
	}
	var exprType ExprType
	switch next {
	case TokenGE:
		exprType = ExprGE

	case TokenGT:
		exprType = ExprGT

	case TokenLE:
		exprType = ExprLE

	case TokenLT:
		exprType = ExprLT

	default:
		return expr1, nil
	}
	_, err = p.getToken()
	if err != nil {
		return nil, err
	}
	right, err := p.getToken()
	if err != nil {
		return nil, err
	}
	expr2, err := p.parseComparison(right)
	if err != nil {
		return nil, err
	}
	return &Expr{
		Type:  exprType,
		Left:  expr1,
		Right: expr2,
	}, nil
}

func (p *Parser) parseAdditive(left *Token) (*Expr, error) {
	expr1, err := p.parseMultiplicative(left)
	if err != nil {
		return nil, err
	}
	next, err := p.peekToken()
	if err != nil {
		return nil, err
	}
	var exprType ExprType
	switch next {
	case TokenPlus:
		exprType = ExprPlus

	case TokenMinus:
		exprType = ExprMinus

	default:
		return expr1, nil
	}
	_, err = p.getToken()
	if err != nil {
		return nil, err
	}
	right, err := p.getToken()
	if err != nil {
		return nil, err
	}
	expr2, err := p.parseAdditive(right)
	if err != nil {
		return nil, err
	}
	return &Expr{
		Type:  exprType,
		Left:  expr1,
		Right: expr2,
	}, nil
}

func (p *Parser) parseMultiplicative(left *Token) (*Expr, error) {
	expr1, err := p.parseLiterals(left)
	if err != nil {
		return nil, err
	}
	next, err := p.peekToken()
	if err != nil {
		return nil, err
	}
	var exprType ExprType
	switch next {
	case TokenMul:
		exprType = ExprMul

	case TokenDiv:
		exprType = ExprDiv

	default:
		return expr1, nil
	}
	_, err = p.getToken()
	if err != nil {
		return nil, err
	}
	right, err := p.getToken()
	if err != nil {
		return nil, err
	}
	expr2, err := p.parseMultiplicative(right)
	if err != nil {
		return nil, err
	}
	return &Expr{
		Type:  exprType,
		Left:  expr1,
		Right: expr2,
	}, nil
}

func (p *Parser) parseLiterals(token *Token) (*Expr, error) {
	switch token.Type {
	case TokenVariable:
		symbol, _ := Intern(token.Value, false)
		return &Expr{
			Type:  ExprVariable,
			Value: NewTermVariable(symbol),
		}, nil

	case TokenIdentifier, TokenString:
		return &Expr{
			Type:  ExprConstant,
			Value: NewTermConstant(token.Value, token.Type == TokenString),
		}, nil

	default:
		return nil, fmt.Errorf("Invalid literal type %s", token)
	}
}
