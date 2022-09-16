package fiqlparser

import (
	"bytes"
	"errors"
	"fmt"
	"strings"
	"unicode"
)

type tokenType int

const tokenValue tokenType = 10
const tokenWildcard tokenType = 11

const tokenBraceOpen tokenType = 20  // (
const tokenBraceClose tokenType = 21 // )
const tokenAND tokenType = 50        //;
const tokenOR tokenType = 51         //,

// standard FIQL comparison
// https://datatracker.ietf.org/doc/html/draft-nottingham-atompub-fiql-00
const tokenCompareEqual = 61    //==
const tokenCompareNotEqual = 62 //!=
const tokenCompareGt = 63       // =gt=
const tokenCompareLt = 64       // =lt=
const tokenCompareGte = 65      // =ge=
const tokenCompareLte = 66      // =le=
//additional comparisons

const tokenComapreIn = 67 // =in=
const tokenCompareQ = 68  // =q= can be used for fuzzy earching etc, q is for query

const tokenEOF = 0

func (t tokenType) String() string {
	switch t {
	case tokenValue:
		return "Value"
	case tokenWildcard:
		return "*"
	case tokenBraceOpen:
		return "("
	case tokenBraceClose:
		return ")"
	case tokenAND:
		return "AND"
	case tokenOR:
		return "OR"
	case tokenCompareEqual:
		return "=="
	case tokenCompareNotEqual:
		return "<>"
	case tokenCompareGt:
		return ">"
	case tokenCompareLt:
		return "<"
	case tokenCompareGte:
		return ">="
	case tokenCompareLte:
		return "<="
	case tokenComapreIn:
		return "in"
	case tokenCompareQ:
		return "query"
	}
	return "eof"
}

func isCompareToken(t tokenType) bool {
	switch t {
	case tokenCompareEqual, tokenCompareNotEqual, tokenCompareGt, tokenCompareLt, tokenCompareGte, tokenCompareLte, tokenComapreIn, tokenCompareQ:
		return true
	}
	return false
}

func isNumberOrDateComparision(t tokenType) bool {
	switch t {
	case tokenCompareGt, tokenCompareLt, tokenCompareGte, tokenCompareLte:
		return true
	}
	return false
}

func isInToken(t tokenType) bool {
	return t == tokenComapreIn
}

func isLogicToken(t tokenType) bool {
	return t == tokenAND || t == tokenOR
}

// ErrUnexpectedInput is generated once unexpected input is met
var ErrUnexpectedInput = errors.New("unexpected input")

// ErrUnexpectedEOF is generated if the input is suspected to be incomplete
var ErrUnexpectedEOF = errors.New("unexpected end of file")

type lexer struct {
	input      []rune
	pos        int
	ln         int
	posInLine  int
	currentVal string
}

func (p *lexer) lastValue() string {
	return p.currentVal
}

func (p *lexer) toCompareToken(cmp string) (tokenType, error) {
	switch strings.ToLower(cmp) {
	case "==":
		return tokenCompareEqual, nil
	case "!=":
		return tokenCompareNotEqual, nil
	case "=gt=":
		return tokenCompareGt, nil
	case "=ge=":
		return tokenCompareGte, nil
	case "=lt=":
		return tokenCompareLt, nil
	case "=le=":
		return tokenCompareLte, nil
	case "=in=":
		return tokenComapreIn, nil
	case "=q=":
		return tokenCompareQ, nil
	}
	return tokenEOF, fmt.Errorf("ln:%d:%d %w (got `%s` but expected one of ==,!=,=gt=,=ge=,=lt=,=le=,=in=,=q=)", p.ln, p.posInLine, ErrUnexpectedInput, cmp)
}

func (p *lexer) readComparator() (tokenType, error) {
	var b bytes.Buffer
	//consume first =
	b.WriteRune(p.consume())
	for {
		r, ok := p.peek()
		if !ok {
			return tokenEOF, ErrUnexpectedEOF
		}
		if r != '=' && r != 'g' && r != 'l' && r != 't' && r != 'e' && r != 'q' && r != 'i' && r != 'n' {
			b.WriteRune(r)
			return tokenEOF, fmt.Errorf("ln:%d:%d %w (got `%s` but expected one of ==,!=,=gt=,=ge=,=lt=,=le=,=in=,=q=)", p.ln, p.posInLine, ErrUnexpectedInput, b.String())
		}
		b.WriteRune(rune(r))
		p.consume()
		if r == '=' {
			break
		}
	}
	return p.toCompareToken(b.String())
}

func (p *lexer) peek() (rune, bool) {
	if p.pos < len(p.input) {
		return p.input[p.pos], true
	}
	return 0, false
}

func (p *lexer) consume() rune {
	// fmt.Printf("Consumed: %c \r\n", p.input[p.pos])
	r := p.input[p.pos]
	if r == '\n' {
		p.ln = p.ln + 1
		p.posInLine = 0
	} else {
		p.posInLine = p.posInLine + 1
	}
	p.pos++
	return r
}

func (p *lexer) readValue() (tokenType, string, error) {
	var b bytes.Buffer
	escaped := false
	c := p.consume()
	if c == '\\' {
		escaped = true
	} else {
		b.WriteRune(c)
	}
	for {
		if p.pos >= len(p.input) {
			break
		}
		v, ok := p.peek()
		if ok {
			if unicode.IsSpace(v) {
				break
			}
			if !escaped && (v == ';' || v == ',' || v == '!' || v == '=' || v == ')' || v == '*') {
				break
			}
		}
		if v == '\\' && !escaped {
			escaped = true
			p.consume()
		} else {
			b.WriteRune(p.consume())
			escaped = false
		}

	}
	val := b.String()
	p.currentVal = val
	return tokenValue, val, nil
}

func (p *lexer) PeekNextToken() (tokenType, string, error) {
	ln := p.ln
	pos := p.pos
	posln := p.posInLine
	val := p.currentVal
	t, err := p.ConsumeToken()
	newCur := p.currentVal
	p.currentVal = val
	p.ln = ln
	p.pos = pos
	p.posInLine = posln
	return t, newCur, err
}

func (p *lexer) ConsumeToken() (tokenType, error) {
	for {
		r, ok := p.peek()
		if !ok {
			return tokenEOF, nil
		}

		if unicode.IsSpace(r) {
			p.consume()
			continue
		}

		if r == '!' || r == '=' {
			return p.readComparator()
		}
		if r == '(' {
			p.consume()
			return tokenBraceOpen, nil
		}
		if r == ')' {
			p.consume()
			return tokenBraceClose, nil
		}
		if r == ';' {
			p.consume()
			return tokenAND, nil
		}
		if r == ',' {
			p.consume()
			return tokenOR, nil
		}
		if r == '*' {
			p.consume()
			return tokenWildcard, nil
		}
		t, _, err := p.readValue()
		return t, err
	}
}
