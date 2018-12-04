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
		_, err = p.getToken()
		if err != nil {
			return nil, err
		}
		return p.parseExpr(token, next)
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

func (p *Parser) parseExpr(left *Token, op TokenType) (*Atom, error) {
	right, err := p.getToken()
	if err != nil {
		return nil, err
	}
	lTerm, err := p.makeExprTerm(left)
	if err != nil {
		return nil, err
	}
	rTerm, err := p.makeExprTerm(right)
	if err != nil {
		return nil, err
	}
	predicate, err := op.Symbol()
	if err != nil {
		return nil, err
	}
	return &Atom{
		Predicate: predicate,
		Terms:     []Term{lTerm, rTerm},
	}, nil
}

func (p *Parser) makeExprTerm(token *Token) (Term, error) {
	switch token.Type {
	case TokenVariable:
		symbol, _ := Intern(token.Value, false)
		return NewTermVariable(symbol), nil

	case TokenIdentifier, TokenString:
		return NewTermConstant(token.Value, token.Type == TokenString), nil

	default:
		return nil, fmt.Errorf("Invalid expression element %s", token)
	}
}
