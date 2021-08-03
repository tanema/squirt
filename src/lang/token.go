package lang

const firstReserved = 257
const endOfStream = -1

const (
	tkAnd rune = iota + firstReserved
	tkAttr
	tkBreak
	tkClass
	tkCleanup
	tkDo
	tkElse
	tkElseif
	tkEnd
	tkFalse
	tkFor
	tkFunction
	tkIf
	tkIn
	tkIsa
	tkNext
	tkNil
	tkOr
	tkReturn
	tkThen
	tkTrue
	tkWhile

	tkSpread
	tkEq
	tkGE
	tkLE
	tkNE
	tkShiftLeft
	tkShiftRight
	tkDecrement
	tkDecrEq
	tkIncrement
	tkIncrEq
	tkEOS
	tkNumber
	tkName
	tkString
	reservedCount = tkWhile - firstReserved + 1
)

var tokens = []string{
	"and",
	"attr",
	"break",
	"class",
	"cleanup",
	"do",
	"else",
	"elseif",
	"end",
	"false",
	"for",
	"func",
	"if",
	"in",
	"isa",
	"next",
	"nil",
	"or",
	"return",
	"then",
	"true",
	"while",

	"...",
	"==",
	">=",
	"<=",
	"!=",
	"<<",
	">>",
	"--",
	"-=",
	"++",
	"+=",
	"<eof>",
	"<number>",
	"<name>",
	"<string>",
}

type token struct {
	t           rune
	numberValue float64
	stringValue string
	loc         [4]int
}

func (tk *token) String() string {
	if tk.t == tkName || tk.t == tkString || tk.t == tkNumber {
		return tk.stringValue
	}
	return runeToStr(tk.t)
}

func runeToStr(t rune) string {
	switch {
	case t == endOfStream:
		return "<eof>"
	case t == '\n' || t == '\r':
		return "<newline>"
	case t < firstReserved:
		return string(t)
	case t < tkEOS:
		return tokens[t-firstReserved]
	}
	return tokens[t-firstReserved]
}
