package lang

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"unicode"
)

type scanner struct {
	file       bool
	source     string
	buffer     bytes.Buffer
	r          io.ByteReader
	current    rune
	lineNumber int
	colNumber  int
}

func newScanner(r io.ByteReader, isFile bool, source string) scanner {
	scn := scanner{
		file:       isFile,
		source:     source,
		r:          r,
		lineNumber: 1,
	}
	scn.advance()
	scn.skipShebangs()
	return scn
}

func isNewLine(c rune) bool { return c == '\n' || c == '\r' }
func isDecimal(c rune) bool { return '0' <= c && c <= '9' }

func (s *scanner) scanError(msg string, data ...interface{}) error {
	return ParseErr{
		file:   s.file,
		source: s.source,
		token:  token{t: s.current, loc: [4]int{s.lineNumber, s.colNumber}},
		msg:    fmt.Sprintf(msg, data...),
	}
}

func (s *scanner) expectedError(expected string) error {
	return s.scanError("expected %v but found %v", expected, runeToStr(s.current))
}

func (s *scanner) tkEOS() token {
	return token{t: tkEOS, loc: [4]int{s.lineNumber, s.colNumber}}
}

func (s *scanner) incrementLineNumber() {
	old := s.current
	if s.advance(); isNewLine(s.current) && s.current != old {
		s.advance()
	}
	s.lineNumber++
	s.colNumber = 1
}

func (s *scanner) advance() {
	if c, err := s.r.ReadByte(); err != nil {
		s.current = endOfStream
	} else {
		s.current = rune(c)
	}
	s.colNumber++
}

func (s *scanner) saveAndAdvance() error {
	if err := s.save(s.current); err != nil {
		return err
	}
	s.advance()
	return nil
}

func (s *scanner) save(c rune) error {
	return s.buffer.WriteByte(byte(c))
}

func (s *scanner) readMultiLine(comment bool) (string, error) {
	if isNewLine(s.current) {
		s.incrementLineNumber()
	}
	for {
		switch s.current {
		case endOfStream:
			return "", s.scanError("unfinished multiline text")
		case '`':
			if comment {
				if err := s.save(s.current); err != nil {
					return "", err
				}
				continue
			}
			s.advance()
			defer s.buffer.Reset()
			str := s.buffer.String()
			if str[len(str)-1] == '\n' {
				str = str[:len(str)-1]
			}
			return str, nil
		case '*':
			if !comment {
				if err := s.save(s.current); err != nil {
					return "", err
				}
				continue
			}
			s.advance()
			if s.current != '/' {
				if err := s.save(s.current); err == nil {
					continue
				} else if err != nil {
					return "", err
				}
			}
			s.advance()
			defer s.buffer.Reset()
			str := s.buffer.String()
			if str[len(str)-1] == '\n' {
				str = str[:len(str)-1]
			}
			return str, nil
		case '\r':
			s.current = '\n'
			fallthrough
		case '\n':
			if err := s.save(s.current); err != nil {
				return "", err
			}
			s.incrementLineNumber()
		default:
			if err := s.save(s.current); err != nil {
				return "", err
			}
			s.advance()
		}
	}
}

func (s *scanner) readDigits() (c rune, e error) {
	for c = s.current; isDecimal(c); c = s.current {
		if e = s.saveAndAdvance(); e != nil {
			return
		}
	}
	return
}

func isHexadecimal(c rune) bool {
	return '0' <= c && c <= '9' || 'a' <= c && c <= 'f' || 'A' <= c && c <= 'F'
}

func (s *scanner) readHexNumber(x float64) (n float64, c rune, i int, err error) {
	if c, n = s.current, x; !isHexadecimal(c) {
		return
	}
	for {
		switch {
		case '0' <= c && c <= '9':
			c = c - '0'
		case 'a' <= c && c <= 'f':
			c = c - 'a' + 10
		case 'A' <= c && c <= 'F':
			c = c - 'A' + 10
		default:
			return
		}
		if err = s.saveAndAdvance(); err != nil {
			return
		}
		c, n, i = s.current, n*16.0+float64(c), i+1
	}
}

func (s *scanner) readNumber() (token, error) {
	const bits64, base10 = 64, 10
	startCol := s.colNumber
	c := s.current
	if err := s.saveAndAdvance(); err != nil {
		return s.tkEOS(), err
	}
	if c == '0' && strings.ContainsRune("Xx", s.current) { // hexadecimal
		s.saveAndAdvance()
		var exponent int
		fraction, c, i, err := s.readHexNumber(0)
		if err != nil {
			return s.tkEOS(), err
		}
		if c == '.' {
			if err := s.saveAndAdvance(); err != nil {
				return s.tkEOS(), err
			}
			fraction, c, exponent, err = s.readHexNumber(fraction)
			if err != nil {
				return s.tkEOS(), err
			}
		}
		if i == 0 && exponent == 0 {
			return s.tkEOS(), s.scanError("malformed number")
		}
		exponent *= -4

		strNum := s.buffer.String()
		s.buffer.Reset()

		if c == 'p' || c == 'P' {
			p := c
			s.advance()
			var negativeExponent bool
			var exp rune
			if c = s.current; c == '+' || c == '-' {
				negativeExponent = c == '-'
				exp = c
				s.advance()
			}
			if !isDecimal(s.current) {
				return s.tkEOS(), s.scanError("malformed number")
			}
			if _, err := s.readDigits(); err != nil {
				return s.tkEOS(), err
			}
			if e, err := strconv.ParseInt(s.buffer.String(), base10, bits64); err != nil {
				return s.tkEOS(), err
			} else if negativeExponent {
				exponent += int(-e)
			} else {
				exponent += int(e)
			}
			strNum = strNum + string(p) + string(exp) + s.buffer.String()
			s.buffer.Reset()
		}
		return token{
			t:           tkNumber,
			numberValue: math.Ldexp(fraction, exponent),
			stringValue: strNum,
			loc:         [4]int{s.lineNumber, startCol, s.lineNumber, s.colNumber - 1},
		}, nil
	}

	c, err := s.readDigits()
	if err != nil {
		return s.tkEOS(), err
	} else if c == '.' {
		if err := s.saveAndAdvance(); err != nil {
			return s.tkEOS(), err
		}
		c, err = s.readDigits()
		if err != nil {
			return s.tkEOS(), err
		}
	}
	if c == 'e' || c == 'E' {
		if err := s.saveAndAdvance(); err != nil {
			return s.tkEOS(), err
		}
		if c = s.current; c == '+' || c == '-' {
			if err := s.saveAndAdvance(); err != nil {
				return s.tkEOS(), err
			}
		}
		_, err = s.readDigits()
		if err != nil {
			return s.tkEOS(), err
		}
	}
	str := s.buffer.String()
	if strings.HasPrefix(str, "0") {
		if str = strings.TrimLeft(str, "0"); str == "" || !isDecimal(rune(str[0])) {
			str = "0" + str
		}
	}
	f, err := strconv.ParseFloat(str, bits64)
	if err != nil {
		return s.tkEOS(), s.scanError("malformed number")
	}
	s.buffer.Reset()
	return token{
		t:           tkNumber,
		numberValue: f,
		stringValue: str,
		loc:         [4]int{s.lineNumber, startCol, s.lineNumber, s.colNumber - 1},
	}, nil
}

var escapes map[rune]rune = map[rune]rune{
	'a': '\a', 'b': '\b', 'f': '\f', 'n': '\n', 'r': '\r', 't': '\t', 'v': '\v', '\\': '\\', '"': '"', '\'': '\'',
}

func (s *scanner) readHexEscape() (r rune, e error) {
	s.advance()
	for i, c, b := 1, s.current, [3]rune{'x'}; i < len(b); i, c, r = i+1, s.current, r<<4+c {
		switch b[i] = c; {
		case '0' <= c && c <= '9':
			c = c - '0'
		case 'a' <= c && c <= 'f':
			c = c - 'a' + 10
		case 'A' <= c && c <= 'F':
			c = c - 'A' + 10
		default:
			return 0, s.expectedError("hexadecimal digit")
		}
		s.advance()
	}
	return
}

func (s *scanner) readDecimalEscape() (r rune, e error) {
	b := [3]rune{}
	for c, i := s.current, 0; i < len(b) && isDecimal(c); i, c = i+1, s.current {
		b[i], r = c, 10*r+c-'0'
		s.advance()
	}
	if r > math.MaxUint8 {
		return 0, s.scanError("decimal escape too large")
	}
	return
}

func (s *scanner) readString() (token, error) {
	delimiter := s.current
	startCol := s.colNumber
	var err error
	for err = s.saveAndAdvance(); err == nil && s.current != delimiter; {
		switch s.current {
		case endOfStream, '\n', '\r':
			return s.tkEOS(), s.scanError("unfinished string")
		case '\\':
			s.advance()
			c := s.current
			switch esc, ok := escapes[c]; {
			case ok:
				s.advance()
				if err := s.save(esc); err != nil {
					return s.tkEOS(), err
				}
			case isNewLine(c):
				s.incrementLineNumber()
				if err := s.save('\n'); err != nil {
					return s.tkEOS(), err
				}
			case c == endOfStream: // do nothing
			case c == 'x':
				if r, err := s.readHexEscape(); err == nil {
					return s.tkEOS(), err
				} else if err = s.save(r); err != nil {
					return s.tkEOS(), err
				}
			case c == 'z':
				for s.advance(); unicode.IsSpace(s.current); {
					if isNewLine(s.current) {
						s.incrementLineNumber()
					} else {
						s.advance()
					}
				}
			default:
				if !isDecimal(c) {
					return s.tkEOS(), s.scanError("invalid escape sequence")
				}
				if r, err := s.readDecimalEscape(); err == nil {
					return s.tkEOS(), err
				} else if err = s.save(r); err != nil {
					return s.tkEOS(), err
				}
			}
		default:
			err = s.saveAndAdvance()
		}
	}
	if err != nil {
		return s.tkEOS(), err
	}
	if err := s.saveAndAdvance(); err != nil {
		return s.tkEOS(), err
	}
	str := s.buffer.String()
	s.buffer.Reset()
	return token{
		t:           tkString,
		stringValue: str[1 : len(str)-1],
		loc:         [4]int{s.lineNumber, startCol, s.lineNumber, s.colNumber},
	}, nil
}

func (s *scanner) skipShebangs() {
	if s.current == '#' {
		if s.advance(); s.current == '!' {
			for !isNewLine(s.current) && s.current != endOfStream {
				s.advance()
			}
			return
		}
	}
}

func (s *scanner) scan() (token, error) {
	const comment, str = true, false
	startLine := s.lineNumber
	startCol := s.colNumber

	for {
		switch c := s.current; c {
		case '\n', '\r':
			s.incrementLineNumber()
		case ' ', '\f', '\t', '\v':
			s.advance()
		case '/':
			s.advance()
			switch s.current {
			case '/':
				for !isNewLine(s.current) && s.current != endOfStream {
					s.advance()
				}
			case '*':
				if _, err := s.readMultiLine(true); err != nil {
					return s.tkEOS(), err
				}
			default:
				return token{
					t:   '/',
					loc: [4]int{s.lineNumber, s.colNumber, s.lineNumber, s.colNumber - 1},
				}, nil
			}
		case '`':
			s.advance()
			strVal, err := s.readMultiLine(false)
			if err != nil {
				return s.tkEOS(), err
			}
			return token{
				t:           tkString,
				stringValue: strVal,
				loc:         [4]int{startLine, startCol, s.lineNumber, s.colNumber},
			}, err
		case '=':
			if s.advance(); s.current != '=' {
				return token{
					t:   '=',
					loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber - 1},
				}, nil
			}
			defer s.advance()
			return token{
				t:   tkEq,
				loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber},
			}, nil
		case '<':
			s.advance()
			if s.current == '=' {
				defer s.advance()
				return token{
					t:   tkLE,
					loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber},
				}, nil
			} else if s.current == '<' {
				defer s.advance()
				return token{
					t:   tkShiftLeft,
					loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber},
				}, nil
			}
			return token{
				t:   '<',
				loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber - 1},
			}, nil
		case '>':
			if s.advance(); s.current == '=' {
				defer s.advance()
				return token{
					t:   tkGE,
					loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber},
				}, nil
			} else if s.current == '>' {
				defer s.advance()
				return token{
					t:   tkShiftRight,
					loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber},
				}, nil
			}
			return token{
				t:   '>',
				loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber - 1},
			}, nil
		case '!':
			if s.advance(); s.current != '=' {
				return token{
					t:   '!',
					loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber - 1},
				}, nil
			}
			defer s.advance()
			return token{
				t:   tkNE,
				loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber},
			}, nil
		case '-':
			if s.advance(); s.current == '-' {
				defer s.advance()
				return token{
					t:   tkDecrement,
					loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber},
				}, nil
			} else if s.current == '=' {
				defer s.advance()
				return token{
					t:   tkDecrEq,
					loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber},
				}, nil
			}
			return token{
				t:   '-',
				loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber - 1},
			}, nil
		case '+':
			if s.advance(); s.current == '+' {
				defer s.advance()
				return token{
					t:   tkIncrement,
					loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber},
				}, nil
			} else if s.current == '=' {
				defer s.advance()
				return token{
					t:   tkIncrEq,
					loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber},
				}, nil
			}
			return token{
				t:   '+',
				loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber - 1},
			}, nil
		case '"', '\'':
			return s.readString()
		case endOfStream:
			return s.tkEOS(), nil
		case '.':
			defer s.buffer.Reset()
			if err := s.saveAndAdvance(); err != nil {
				return s.tkEOS(), err
			}
			if s.current == '.' {
				if s.advance(); s.current == '.' {
					defer s.advance()
					return token{
						t:   tkSpread,
						loc: [4]int{s.lineNumber, s.colNumber - 2, s.lineNumber, s.colNumber},
					}, nil
				}
				return s.tkEOS(), s.scanError("unexpected token .. found")
			} else if !unicode.IsDigit(s.current) {
				s.buffer.Reset()
				return token{
					t:   '.',
					loc: [4]int{s.lineNumber, s.colNumber, s.lineNumber, s.colNumber},
				}, nil
			} else {
				tk, err := s.readNumber()
				if err == nil {
					tk.loc[1] = tk.loc[1] - 1
				}
				return tk, err
			}
		case 0:
			s.advance()
		default:
			if unicode.IsDigit(c) {
				return s.readNumber()
			} else if c == '_' || unicode.IsLetter(c) {
				for ; c == '_' || unicode.IsLetter(c) || unicode.IsDigit(c); c = s.current {
					if err := s.saveAndAdvance(); err != nil {
						return s.tkEOS(), err
					}
				}
				str := s.buffer.String()
				s.buffer.Reset()
				for i, reserved := range tokens[:reservedCount] {
					if str == reserved {
						return token{
							t:           rune(i + firstReserved),
							stringValue: reserved,
							loc:         [4]int{startLine, startCol, s.lineNumber, s.colNumber - 1},
						}, nil
					}
				}
				return token{
					t:           tkName,
					stringValue: str,
					loc:         [4]int{s.lineNumber, s.colNumber - len(str), s.lineNumber, s.colNumber - 1},
				}, nil
			}
			s.advance()
			return token{
				t:   c,
				loc: [4]int{s.lineNumber, s.colNumber - 1, s.lineNumber, s.colNumber - 1},
			}, nil
		}
	}
}
