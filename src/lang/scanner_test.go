package lang

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type scannertest struct {
	source string
	tokens []token
	err    string
}

func TestScanner(t *testing.T) {
	tests := []scannertest{
		// smoke
		{source: ""},
		// comments
		{source: "// short comment"},
		{source: "/* long comment */"},
		{source: "/* long *com/ment */"},
		{source: "// hello, world\n"},
		// simple tokens
		{source: "-", tokens: []token{{t: '-'}}},
		{source: "=", tokens: []token{{t: '='}}},
		{source: ".", tokens: []token{{t: '.'}}},
		{source: ":", tokens: []token{{t: ':'}}},
		// strings
		{source: "\"hello, world\"", tokens: []token{{t: tkString, stringValue: "hello, world"}}},
		{source: "`hello,\r\nworld`", tokens: []token{{t: tkString, stringValue: "hello,\n\nworld"}}},
		{source: "`hello ],\r\nworld`", tokens: []token{{t: tkString, stringValue: "hello ],\n\nworld"}}},
		{source: "`hello world", err: "unfinished multiline text"},
		// names
		{source: "_foo", tokens: []token{{t: tkName, stringValue: "_foo"}}},
		{source: "baz123", tokens: []token{{t: tkName, stringValue: "baz123"}}},
		{source: "boo_boo", tokens: []token{{t: tkName, stringValue: "boo_boo"}}},
		// keywords and composite tokens
		{source: "..", err: "unexpected token .. found"},
		{source: "and", tokens: []token{{t: tkAnd, stringValue: "and"}}},
		{source: "attr", tokens: []token{{t: tkAttr, stringValue: "attr"}}},
		{source: "break", tokens: []token{{t: tkBreak, stringValue: "break"}}},
		{source: "do", tokens: []token{{t: tkDo, stringValue: "do"}}},
		{source: "else", tokens: []token{{t: tkElse, stringValue: "else"}}},
		{source: "elseif", tokens: []token{{t: tkElseif, stringValue: "elseif"}}},
		{source: "end", tokens: []token{{t: tkEnd, stringValue: "end"}}},
		{source: "false", tokens: []token{{t: tkFalse, stringValue: "false"}}},
		{source: "func", tokens: []token{{t: tkFunction, stringValue: "func"}}},
		{source: "if", tokens: []token{{t: tkIf, stringValue: "if"}}},
		{source: "for", tokens: []token{{t: tkFor, stringValue: "for"}}},
		{source: "in", tokens: []token{{t: tkIn, stringValue: "in"}}},
		{source: "nil", tokens: []token{{t: tkNil, stringValue: "nil"}}},
		{source: "or", tokens: []token{{t: tkOr, stringValue: "or"}}},
		{source: "next", tokens: []token{{t: tkNext, stringValue: "next"}}},
		{source: "return", tokens: []token{{t: tkReturn, stringValue: "return"}}},
		{source: "then", tokens: []token{{t: tkThen, stringValue: "then"}}},
		{source: "true", tokens: []token{{t: tkTrue, stringValue: "true"}}},
		{source: "while", tokens: []token{{t: tkWhile, stringValue: "while"}}},
		{source: "...", tokens: []token{{t: tkSpread}}},
		{source: "==", tokens: []token{{t: tkEq}}},
		{source: ">=", tokens: []token{{t: tkGE}}},
		{source: "<=", tokens: []token{{t: tkLE}}},
		{source: "!=", tokens: []token{{t: tkNE}}},
		{source: "<<", tokens: []token{{t: tkShiftLeft}}},
		{source: ">>", tokens: []token{{t: tkShiftRight}}},
		{source: "--", tokens: []token{{t: tkDecrement}}},
		{source: "-=", tokens: []token{{t: tkDecrEq}}},
		{source: "++", tokens: []token{{t: tkIncrement}}},
		{source: "+=", tokens: []token{{t: tkIncrEq}}},
		{source: "class", tokens: []token{{t: tkClass, stringValue: "class"}}},
		{source: "isa", tokens: []token{{t: tkIsa, stringValue: "isa"}}},
		// numbers
		{source: ".34", tokens: []token{{t: tkNumber, numberValue: 0.34, stringValue: ".34"}}},
		{source: "3", tokens: []token{{t: tkNumber, numberValue: float64(3), stringValue: "3"}}},
		{source: "3.0", tokens: []token{{t: tkNumber, numberValue: 3.0, stringValue: "3.0"}}},
		{source: "3.1416", tokens: []token{{t: tkNumber, numberValue: 3.1416, stringValue: "3.1416"}}},
		{source: "3.1416e", err: "malformed number"},
		{source: "314.16e-2", tokens: []token{{t: tkNumber, numberValue: 3.1416, stringValue: "314.16e-2"}}},
		{source: "0.31416E1", tokens: []token{{t: tkNumber, numberValue: 3.1416, stringValue: "0.31416E1"}}},
		{source: "0xff", tokens: []token{{t: tkNumber, numberValue: float64(0xff), stringValue: "0xff"}}},
		{source: "0x0.1E", tokens: []token{{t: tkNumber, numberValue: 0.1171875, stringValue: "0x0.1E"}}},
		{source: "0xA23p-4", tokens: []token{{t: tkNumber, numberValue: 162.1875, stringValue: "0xA23p-4"}}},
		{source: "0xA23p", err: "malformed number"},
		{source: "0X1.921FB54442D18P+1", tokens: []token{{t: tkNumber, numberValue: 3.141592653589793, stringValue: "0X1.921FB54442D18P+1"}}},
		{source: "  -0xa  ", tokens: []token{{t: '-'}, {t: tkNumber, numberValue: 10.0, stringValue: "0xa"}}},
	}
	for n, v := range tests {
		s := scanner{r: strings.NewReader(v.source)}
		actual := []token{}
		result, err := s.scan()
		for result.t != tkEOS && err == nil {
			actual = append(actual, result)
			result, err = s.scan()
		}
		assert.Equal(t, result.t, tkEOS)
		if assert.Equal(t, len(v.tokens), len(actual), fmt.Sprintf("[%d] wrong number of received tokens", n)) {
			for i, expected := range v.tokens {
				assert.Equal(t, expected.t, actual[i].t, fmt.Sprintf("[%d] wrong token '%v' != '%v'", n, expected.stringValue, actual[i].stringValue))
				assert.Equal(t, expected.stringValue, actual[i].stringValue)
				assert.Equal(t, expected.numberValue, actual[i].numberValue)
			}
		} else {
			for _, a := range actual {
				println(tokDebug(a))
			}
		}
		if v.err != "" {
			assert.NotNil(t, err)
		} else {
			assert.Nil(t, err)
		}
		if err != nil {
			assert.Contains(t, err.Error(), v.err)
		}
	}
}

func tokDebug(t token) string {
	tok := string(t.t)
	if tkAnd <= t.t && t.t <= tkString {
		tok = tokens[t.t-firstReserved]
	}
	return fmt.Sprintf("{t:%s, n:%f, s:%q}", tok, t.numberValue, t.stringValue)
}
