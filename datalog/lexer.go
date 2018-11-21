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

type Lexer struct {
	in      *bufio.Reader
	current *Position
	last    *Position
}

type Position struct {
	Name string
	Row  int
	Col  int
}

func (p Position) String() string {
	return fmt.Sprintf("%s:%d:%d", p.Name, p.Row, p.Col)
}

func (p *Position) Set(o *Position) {
	p.Name = o.Name
	p.Row = o.Row
	p.Col = o.Col
}

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

func (l *Lexer) ReadRune() (rune, error) {
	r, n, err := l.in.ReadRune()
	if err != nil {
		return r, err
	}
	l.last.Set(l.current)
	if r == '\n' {
		l.current.Row++
		l.current.Col = 0
	} else {
		l.current.Col += n
	}

	return r, err
}

func (l *Lexer) UnreadRune() {
	l.in.UnreadRune()
	l.current.Set(l.last)
}

func (l *Lexer) Pos() string {
	return l.current.String()
}

func (l *Lexer) GetToken() (*Token, error) {
	for {
		r, err := l.ReadRune()
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

		case '(', ',', ')', '=', '.', '~', '?':
			return &Token{
				Type:     TokenType(r),
				Position: *l.last,
			}, nil

		case ':':
			r, err := l.ReadRune()
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
		r, err := l.ReadRune()
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
		r, err := l.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if unicode.IsLetter(r) || unicode.IsNumber(r) || r == '_' {
			value = append(value, r)
		} else {
			l.UnreadRune()
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
		r, err := l.ReadRune()
		if err != nil {
			return nil, err
		}
		if r == '"' {
			break
		}
		if r == '\\' {
			r, err := l.ReadRune()
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

func stringify(val string) string {
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
		r, err := l.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if isIdentifierRune(r) {
			value = append(value, r)
		} else {
			l.UnreadRune()
			break
		}
	}
	return &Token{
		Type:     TokenIdentifier,
		Value:    string(value),
		Position: *l.last,
	}, nil
}

func isIdentifierRune(r rune) bool {
	switch r {
	case '(', ',', ')', '=', ':', '.', '~', '?', '"', '%':
		return false

	default:
		if unicode.IsSpace(r) {
			return false
		}
	}
	return unicode.IsLetter(r) || unicode.IsNumber(r) || unicode.IsPunct(r)
}

type TokenType int

const (
	TokenArrow TokenType = iota + 256
	TokenError
	TokenVariable
	TokenIdentifier
	TokenString
)

type Token struct {
	Type     TokenType
	Value    string
	Position Position
}

func (t *Token) String() string {
	if t.Type < 256 {
		return fmt.Sprintf("%c", rune(t.Type))
	} else if t.Type == TokenArrow {
		return ":-"
	} else if t.Type == TokenError {
		return "{error}"
	} else {
		return t.Value
	}
}
