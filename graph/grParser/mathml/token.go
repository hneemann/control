package mathml

import (
	"fmt"
	"unicode"
	"unicode/utf8"
)

type Tokenizer struct {
	s       string
	isRune  bool
	rune    rune
	isToken bool
	token   Token
}

type Kind int

func (k Kind) String() string {
	switch k {
	case EOF:
		return "EOF"
	case Number:
		return "Number"
	case Identifier:
		return "Identifier"
	case Command:
		return "Command"
	case Operator:
		return "Operator"
	case OpenParen:
		return "OpenParen"
	case CloseParen:
		return "CloseParen"
	case OpenBrace:
		return "OpenBrace"
	case CloseBrace:
		return "CloseBrace"
	case Ampersand:
		return "Ampersand"
	case Dollar:
		return "Dollar"
	case Linefeed:
		return "Linefeed"
	case Up:
		return "Up"
	case Down:
		return "Down"
	}
	return fmt.Sprintf("Kind(%d)", k)
}

const (
	EOF Kind = iota
	Number
	Identifier
	Command
	Operator
	OpenParen
	CloseParen
	OpenBrace
	CloseBrace
	Ampersand
	Dollar
	Linefeed
	Up
	Down
)

type Token struct {
	kind  Kind
	value string
}

func (t Token) String() string {
	return fmt.Sprintf("%v(%v)", t.kind, t.value)
}

func NewTokenizer(s string) *Tokenizer {
	return &Tokenizer{s: s}
}

func (t *Tokenizer) Peek() rune {
	if t.isRune {
		return t.rune
	}
	if t.s == "" {
		return 0
	}
	var size int
	t.rune, size = utf8.DecodeRuneInString(t.s)
	t.s = t.s[size:]
	t.isRune = true
	return t.rune
}

func (t *Tokenizer) Read() rune {
	r := t.Peek()
	t.isRune = false
	return r
}

func (t *Tokenizer) PeekToken() Token {
	if t.isToken {
		return t.token
	}
	t.token = t.readToken()
	t.isToken = true
	return t.token
}

func (t *Tokenizer) NextToken() Token {
	if t.isToken {
		t.isToken = false
		return t.token
	}
	return t.readToken()
}

func (t *Tokenizer) readToken() Token {
	for {
		switch r := t.Peek(); {
		case unicode.IsSpace(r):
			t.Read()
		case r == 0:
			return Token{kind: EOF}
		case r == '(':
			t.Read()
			return Token{kind: OpenParen, value: "("}
		case r == ')':
			t.Read()
			return Token{kind: CloseParen, value: ")"}
		case r == '{':
			t.Read()
			return Token{kind: OpenBrace, value: "{"}
		case r == '}':
			t.Read()
			return Token{kind: CloseBrace, value: "}"}
		case r == '^':
			t.Read()
			return Token{kind: Up, value: "^"}
		case r == '_':
			t.Read()
			return Token{kind: Down, value: "_"}
		case r == '&':
			t.Read()
			return Token{kind: Ampersand, value: "&"}
		case r == '$':
			t.Read()
			return Token{kind: Dollar, value: "$"}
		case unicode.IsNumber(r):
			var value string
			for r := t.Peek(); unicode.IsNumber(r) || r == '.'; r = t.Peek() {
				value += string(t.Read())
			}
			return Token{kind: Number, value: value}
		case unicode.IsLetter(r):
			var value string
			for r := t.Peek(); unicode.IsLetter(r); r = t.Peek() {
				value += string(t.Read())
			}
			return Token{kind: Identifier, value: value}
		case r == '\\':
			t.Read()
			if t.Peek() == '\\' {
				t.Read()
				return Token{kind: Linefeed, value: "\\\\"}
			} else {
				var value string
				for r := t.Peek(); unicode.IsLetter(r); r = t.Peek() {
					value += string(t.Read())
				}
				return Token{kind: Command, value: value}
			}
		default:
			t.Read()
			return Token{kind: Operator, value: string(r)}
		}
	}
}
