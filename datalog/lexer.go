//
// lexer.go
//
// Copyright (c) 2018 Markku Rossi
//
// All rights reserved.
//

package datalog

import (
	"bufio"
	"fmt"
	"io"
	"unicode"
)

// Lexer implements lexical analyzer.
type Lexer struct {
	in      *bufio.Reader
	current *Position
	last    *Position
}

// Position implements input positions.
type Position struct {
	Name string
	Row  int
	Col  int
}

func (p Position) String() string {
	return fmt.Sprintf("%s:%d:%d", p.Name, p.Row, p.Col)
}

// Set sets the position.
func (p *Position) Set(o *Position) {
	p.Name = o.Name
	p.Row = o.Row
	p.Col = o.Col
}

// NewLexer creates a new lexical analyzer.
func NewLexer(inputName string, rd io.Reader) *Lexer {
	return &Lexer{
		in: bufio.NewReader(rd),
		current: &Position{
			Name: inputName,
			Row:  1,
		},
		last: &Position{},
	}
}

// ReadRune reads the next input rune.
func (l *Lexer) ReadRune() (rune, int, error) {
	r, n, err := l.in.ReadRune()
	if err != nil {
		return r, n, err
	}
	l.last.Set(l.current)
	if r == '\n' {
		l.current.Row++
		l.current.Col = 0
	} else {
		l.current.Col += n
	}

	return r, n, err
}

// UnreadRune unreads the latest input rune.
func (l *Lexer) UnreadRune() error {
	err := l.in.UnreadRune()
	if err != nil {
		return err
	}
	l.current.Set(l.last)
	return nil
}

// Pos returns the current input position as a string.
func (l *Lexer) Pos() string {
	return l.current.String()
}

// GetToken gets the next token.
func (l *Lexer) GetToken() (*Token, error) {
	for {
		r, _, err := l.ReadRune()
		if err != nil {
			return nil, err
		}
		if unicode.IsSpace(r) {
			continue
		}
		switch r {
		case '%':
			if err = l.skipComment(); err != nil {
				return nil, err
			}
			continue

		case '=':
			return &Token{
				Type:     TokenEQ,
				Position: *l.last,
			}, nil

		case '>':
			r, _, err := l.ReadRune()
			t := TokenGT
			if err == nil {
				if r == '=' {
					t = TokenGE
				} else {
					if err := l.UnreadRune(); err != nil {
						return nil, err
					}
				}
			}
			return &Token{
				Type:     t,
				Position: *l.last,
			}, nil

		case '<':
			r, _, err := l.ReadRune()
			t := TokenLT
			if err == nil {
				if r == '=' {
					t = TokenLE
				} else {
					if err := l.UnreadRune(); err != nil {
						return nil, err
					}
				}
			}
			return &Token{
				Type:     t,
				Position: *l.last,
			}, nil

		case '*':
			return &Token{
				Type:     TokenMul,
				Position: *l.last,
			}, nil

		case '/':
			return &Token{
				Type:     TokenDiv,
				Position: *l.last,
			}, nil

		case '+':
			return &Token{
				Type:     TokenPlus,
				Position: *l.last,
			}, nil

		case '-':
			return &Token{
				Type:     TokenMinus,
				Position: *l.last,
			}, nil

		case '(', ',', ')', '.', '~', '?':
			return &Token{
				Type:     TokenType(r),
				Position: *l.last,
			}, nil

		case ':':
			r, _, err := l.ReadRune()
			if err != nil {
				return nil, err
			}
			if r == '-' {
				return &Token{
					Type:     TokenArrow,
					Position: *l.last,
				}, nil
			}
			return nil, fmt.Errorf("%s: Invalid input: %v", l.Pos(), r)

		default:
			if unicode.IsUpper(r) {
				return l.readVariable(r)
			}
			if r == '"' {
				return l.readString()
			}
			return l.readIdentifier(r)
		}
	}
}

func (l *Lexer) skipComment() error {
	for {
		r, _, err := l.ReadRune()
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return err
		}
		if r == '\n' {
			return nil
		}
	}
}

func (l *Lexer) readVariable(r rune) (*Token, error) {
	value := []rune{r}
	for {
		r, _, err := l.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
			value = append(value, r)
		} else {
			if err := l.UnreadRune(); err != nil {
				return nil, err
			}
			break
		}
	}

	return &Token{
		Type:     TokenVariable,
		Value:    string(value),
		Position: *l.last,
	}, nil
}

func (l *Lexer) readString() (*Token, error) {
	var value []rune
	for {
		r, _, err := l.ReadRune()
		if err != nil {
			return nil, err
		}
		if r == '"' {
			break
		}
		if r == '\\' {
			r, _, err = l.ReadRune()
			if err != nil {
				return nil, err
			}
			switch r {
			case '\n':
				// Line continuation
				continue

			case 'a':
				r = '\a'
			case 'b':
				r = '\b'
			case 'f':
				r = '\f'
			case 'n':
				r = '\n'
			case 'r':
				r = '\r'
			case 't':
				r = '\t'
			case 'v':
				r = '\v'
			case '\'':
				r = '\''
			case '?':
				r = '?'
				// XXX octal and other character escapes.
			}
		}
		value = append(value, r)
	}
	return &Token{
		Type:     TokenString,
		Value:    string(value),
		Position: *l.last,
	}, nil
}

// Stringify escapes the argument string so that it is a valid datalog
// string literal value.
func Stringify(val string) string {
	result := []rune{'"'}
	for _, r := range val {
		switch r {
		case '\a':
			result = append(result, []rune{'\\', 'a'}...)
		case '\b':
			result = append(result, []rune{'\\', 'b'}...)
		case '\f':
			result = append(result, []rune{'\\', 'f'}...)
		case '\n':
			result = append(result, []rune{'\\', 'n'}...)
		case '\r':
			result = append(result, []rune{'\\', 'r'}...)
		case '\t':
			result = append(result, []rune{'\\', 't'}...)
		case '\v':
			result = append(result, []rune{'\\', 'v'}...)
		case '"':
			result = append(result, []rune{'\\', '"'}...)
		default:
			result = append(result, r)
		}
	}
	result = append(result, '"')
	return string(result)
}

func (l *Lexer) readIdentifier(r rune) (*Token, error) {
	value := []rune{r}
	for {
		r, _, err := l.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if isIdentifierRune(r) {
			value = append(value, r)
		} else {
			if err := l.UnreadRune(); err != nil {
				return nil, err
			}
			break
		}
	}
	str := string(value)
	if str == "_" {
		return &Token{
			Type:     TokenWildcard,
			Position: *l.last,
		}, nil
	}
	return &Token{
		Type:     TokenIdentifier,
		Value:    str,
		Position: *l.last,
	}, nil
}

func isIdentifierRune(r rune) bool {
	switch r {
	case '(', ',', ')', ':', '.', '~', '?', '"', '%', '*', '/', '-':
		return false

	default:
		if unicode.IsSpace(r) {
			return false
		}
	}
	return unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsPunct(r)
}

// TokenType defines token types.
type TokenType int

// Known token types.
const (
	TokenArrow TokenType = iota + 256
	TokenWildcard
	TokenEQ
	TokenGE
	TokenGT
	TokenLE
	TokenLT
	TokenMul
	TokenDiv
	TokenPlus
	TokenMinus
	TokenError
	TokenVariable
	TokenIdentifier
	TokenString
)

// IsExpr tests if the token is an expression.
func (t TokenType) IsExpr() bool {
	switch t {
	case TokenEQ, TokenGE, TokenGT, TokenLE, TokenLT, TokenMul, TokenDiv,
		TokenPlus, TokenMinus:
		return true
	default:
		return false
	}
}

// Token implements a datalog program token.
type Token struct {
	Type     TokenType
	Value    string
	Position Position
}

var tokenNames = map[TokenType]string{
	TokenArrow:    ":-",
	TokenWildcard: "_",
	TokenEQ:       "=",
	TokenGE:       ">=",
	TokenGT:       ">",
	TokenLE:       "<=",
	TokenLT:       "<",
	TokenMul:      "*",
	TokenDiv:      "/",
	TokenPlus:     "+",
	TokenMinus:    "-",
	TokenError:    "{error}",
}

func (t *Token) String() string {
	if t.Type < 256 {
		return fmt.Sprintf("%c", rune(t.Type))
	}
	name, ok := tokenNames[t.Type]
	if ok {
		return name
	}
	return t.Value

}
