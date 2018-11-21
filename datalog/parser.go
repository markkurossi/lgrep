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
	clause = &Clause{
		Head: atom,
	}

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
	if token.Type != TokenIdentifier && token.Type != TokenString {
		return nil, fmt.Errorf("%s: unexpected token: %s",
			token.Position, token)
	}

	atom := &Atom{
		Predicate: intern(token.Value, token.Type == TokenString),
	}

	next, err := p.peekToken()
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
				term = NewTermVariable(intern(token.Value, false))

			case TokenIdentifier, TokenString:
				term = NewTermConstant(token.Value)

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
	}

	return atom, nil
}
