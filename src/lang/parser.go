package lang

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"unicode"
)

var binaryPrecedence = map[string]int{
	"^": 12,
	"*": 10, "/": 10, "%": 10,
	"+": 9, "-": 9,
	">>": 7, "<<": 7,
	"&": 6,
	"~": 5,
	"|": 4,
	"<": 3, ">": 3, ">=": 3, "<=": 3, "==": 3, "!=": 3, "++": 3, "--": 3, "+=": 3, "-=": 3,
	"and": 2,
	"or":  1,
}

type parser struct {
	file            bool
	source          string
	scn             scanner
	prev, tk, ahead token
	inLoop          int
	inClass         int
	locations       [][4]int
}

func ParseFile(filepath string) (Object, error) {
	f, err := os.Open(filepath)
	if err != nil {
		return invalid, err
	}
	p := parser{
		file:   true,
		source: filepath,
		scn:    newScanner(bufio.NewReader(f), true, filepath),
		ahead:  token{t: tkEOS},
	}
	return p.parse()
}

func ParseStr(source string) (Object, error) {
	p := parser{
		file:   false,
		source: source,
		scn:    newScanner(strings.NewReader(source), false, source),
		ahead:  token{t: tkEOS},
	}
	return p.parse()
}

func (p *parser) parse() (Object, error) {
	if err := p.next(); err != nil {
		return Object{}, err
	}
	p.pushLoc()
	statements, catches, err := p.block()
	if err != nil {
		return Object{}, err
	}
	return Object{Kind: Root, Block: statements, Catches: catches, Pos: p.popLoc()}, nil
}

func (p *parser) parseError(msg string, data ...interface{}) error {
	return ParseErr{
		file:   p.file,
		source: p.source,
		token:  p.tk,
		msg:    fmt.Sprintf(msg, data...),
	}
}

func (p *parser) expectedErr(expected string) error {
	return p.parseError("expected %v but found %v", expected, runeToStr(p.tk.t))
}

func (p *parser) pushLoc() {
	p.locations = append(p.locations, p.tk.loc)
}

func (p *parser) popLoc() [4]int {
	if len(p.locations) == 0 {
		panic("unbalanced location tracking")
	}
	loc := p.locations[len(p.locations)-1]
	p.locations = p.locations[:len(p.locations)-1]
	return p.endLoc(loc)
}

func (p *parser) endLoc(loc [4]int) [4]int {
	loc[2] = p.prev.loc[2]
	loc[3] = p.prev.loc[3]
	return loc
}

func (p *parser) next() (err error) {
	p.prev = p.tk
	if p.ahead.t != tkEOS {
		p.tk = p.ahead
		p.ahead.t = tkEOS
	} else {
		p.tk, err = p.scn.scan()
	}
	return
}

func (p *parser) lookAhead() (token, error) {
	var err error
	if p.ahead.t == tkEOS {
		p.ahead, err = p.scn.scan()
	}
	return p.ahead, err
}

func (p *parser) nextIf(t rune) error {
	if t == p.tk.t {
		return p.next()
	}
	return nil
}

func (p *parser) expect(t rune) error {
	if p.tk.t == t {
		return p.next()
	}
	return p.expectedErr(runeToStr(t))
}

func (p *parser) block() ([]Object, []Object, error) {
	var err error
	statements := []Object{}
	for !p.isBlockFollow() {
		var statement Object
		switch p.tk.t {
		case ';':
			continue
		case tkIf:
			statement, err = p.ifStatement()
		case tkDo:
			statement, err = p.doStatement()
		case tkFor:
			statement, err = p.forStatement()
		case tkWhile:
			statement, err = p.whileStatement()
		case tkFunction:
			statement, err = p.functionDeclaration()
		case tkReturn:
			statement, err = p.returnStatement()
		case tkBreak:
			statement, err = p.breakStatement()
		case tkNext:
			statement, err = p.nextStatement()
		case tkClass:
			statement, err = p.classStatement()
		default:
			statement, err = p.assignmentOrCallStatement()
		}
		if err != nil {
			return statements, []Object{}, err
		}
		if err := p.nextIf(';'); err != nil {
			return statements, []Object{}, err
		}
		statements = append(statements, statement)
	}

	catches, err := p.consumecleanupStatement()
	return statements, catches, err
}

func (p *parser) isBlockFollow() bool {
	switch p.tk.t {
	case tkEOS, tkElseif, tkElse, tkEnd, tkCleanup:
		return true
	default:
		return false
	}
}

func (p *parser) breakStatement() (Object, error) {
	if p.inLoop <= 0 {
		return invalid, p.parseError("use of a break statement outside of a loop. break can only be used to effect iteration with for, while, and repeat loops")
	}
	pos := p.tk.loc
	if err := p.nextIf(tkBreak); err != nil {
		return invalid, err
	}
	return Object{Kind: Break, Pos: pos}, nil
}

func (p *parser) nextStatement() (Object, error) {
	if p.inLoop <= 0 {
		return invalid, p.parseError("use of a next statement outside of a loop. next can only be used to effect iteration with for, while, and repeat loops")
	}
	pos := p.tk.loc
	if err := p.nextIf(tkNext); err != nil {
		return invalid, err
	}
	return Object{Kind: Next, Pos: pos}, nil
}

func (p *parser) classStatement() (Object, error) {
	p.inClass++
	defer func() { p.inClass-- }()
	p.pushLoc()
	if err := p.nextIf(tkClass); err != nil {
		return invalid, err
	}
	parentClass := Object{}
	className, err := p.identifier()
	if err != nil {
		return invalid, err
	}
	if p.tk.t == tkIsa {
		if err := p.next(); err != nil {
			return invalid, err
		}
		parentClass, err = p.identifier()
		if err != nil {
			return invalid, err
		}
	}

	if err := p.nextIf(tkDo); err != nil {
		return invalid, err
	}

	statements := []Object{}
	for p.tk.t != tkEnd {
		var statement Object
		switch p.tk.t {
		case ';':
			continue
		case tkFunction:
			statement, err = p.functionDeclaration()
			if statement.Value.Kind != Identifier {
				return invalid, p.parseError("non identifier func name in class definition")
			}
			statement.Private, statement.Static = nameVals(statement.Value.Name)
		case tkClass:
			statement, err = p.classStatement()
		case tkAttr:
			statement, err = p.attrDeclaration()
		default:
			return invalid, p.parseError("unexpected statement in class definition")
		}
		if err != nil {
			return invalid, err
		}
		if err := p.nextIf(';'); err != nil {
			return invalid, err
		}
		statements = append(statements, statement)
	}

	if err := p.nextIf(tkEnd); err != nil {
		return invalid, err
	}

	return Object{
		Kind:   ClassDef,
		Name:   className.Name,
		Parent: parentClass.Name,
		Pos:    p.popLoc(),
		Block:  statements,
	}, nil
}

func nameVals(name string) (private, static bool) {
	private = strings.HasPrefix(name, "_")
	static = (private && unicode.IsUpper(rune(name[1]))) || unicode.IsUpper(rune(name[0]))
	return
}

func (p *parser) attrDeclaration() (Object, error) {
	if err := p.nextIf(tkAttr); err != nil {
		return invalid, err
	}
	name, err := p.identifier()
	if err != nil {
		return invalid, err
	}
	private, static := nameVals(name.Name)
	var val *Object
	var refine *Object
	if p.tk.t == '=' {
		if err := p.next(); err != nil {
			return invalid, err
		}
		v, err := p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		val = &v
		if p.tk.t == ',' {
			if err := p.next(); err != nil {
				return invalid, err
			}
			refinement, err := p.tableConstructor()
			if err != nil {
				return invalid, err
			}
			refine = &refinement
		}
	}
	return Object{
		Kind:    AttrDef,
		Name:    name.Name,
		Value:   val,
		Cond:    refine,
		Private: private,
		Static:  static,
	}, nil
}

func (p *parser) assignmentOrCallStatement() (Object, error) {
	var lvalue bool
	var base Object
	var err error
	var targets = []Object{}
	startLoc := p.tk.loc

	for {
		if p.tk.t == tkName {
			base, err = p.identifier()
			if err != nil {
				return invalid, err
			}
			lvalue = true
		} else {
			return p.expectedExpression()
		}

	prefixExprLoop:
		for {
			switch p.tk.t {
			case '.', '[': //indexing values on lvalue
				lvalue = true
			case '(': // function call no longer lvalue
				lvalue = false
			default:
				break prefixExprLoop
			}
			base, err = p.prefixExpressionPart(base)
			if err != nil {
				return invalid, err
			}
		}

		targets = append(targets, base)
		if p.tk.t != ',' {
			break
		}

		if !lvalue {
			return invalid, p.expectedErr(",")
		}

		if err := p.next(); err != nil {
			return invalid, err
		}
	}

	if len(targets) == 1 && !lvalue {
		return targets[0], nil
	} else if lvalue && p.isAssignShortcut(p.tk) {
		if len(targets) > 1 {
			return invalid, p.parseError("cannot use assignment shortcut '%v' with multiple destinations", runeToStr(p.tk.t))
		}
		return p.shortcutAssign(targets[0])
	} else if !lvalue {
		return invalid, p.parseError("not all assignment targets are assignable")
	}

	if p.tk.t != '=' && len(targets) == 1 {
		return targets[0], nil
	} else if err := p.expect('='); err != nil {
		return invalid, err
	}

	var values = []Object{}

	for {
		expr, err := p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		values = append(values, expr)
		if p.tk.t != ',' {
			break
		} else if err := p.next(); err != nil {
			return invalid, err
		}
	}

	return Object{
		Kind: Assignment,
		Vars: targets,
		Vals: values,
		Pos:  p.endLoc(startLoc),
	}, nil
}

func (p *parser) identifier() (Object, error) {
	if p.tk.t != tkName {
		return Object{}, p.expectedErr("<name>")
	}
	idnt := Object{Kind: Identifier, Name: p.tk.stringValue, Pos: p.tk.loc}
	return idnt, p.next()
}

func (p *parser) shortcutAssign(target Object) (Object, error) {
	var value Object
	p.pushLoc()

	operator := p.tk

	if err := p.next(); err != nil {
		return invalid, err
	}

	switch operator.t {
	case tkDecrEq:
		expr, err := p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		value = Object{Kind: Binary, Name: "-", Vals: []Object{target, expr}, Pos: operator.loc}
	case tkDecrement:
		value = Object{Kind: Binary, Name: "-", Vals: []Object{target, {Kind: Number, NumberValue: 1}}, Pos: operator.loc}
	case tkIncrEq:
		expr, err := p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		value = Object{Kind: Binary, Name: "+", Vals: []Object{target, expr}, Pos: operator.loc}
	case tkIncrement:
		value = Object{Kind: Binary, Name: "+", Vals: []Object{target, {Kind: Number, NumberValue: 1}}, Pos: operator.loc}
	case tkShiftLeft:
		expr, err := p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		value = Object{Kind: Binary, Name: "<<", Vals: []Object{target, expr}, Pos: operator.loc}
	}

	return Object{
		Kind: Assignment,
		Vars: []Object{target},
		Vals: []Object{value},
		Pos:  p.popLoc(),
	}, nil
}

func (p *parser) expectedExpression() (Object, error) {
	expression, err := p.expression()
	if err != nil {
		return invalid, err
	} else if expression.Kind == Invalid {
		return invalid, p.expectedErr("expression")
	}
	return expression, nil
}

func (p *parser) expression() (Object, error) { return p.subExpression(0) }

func (p *parser) subExpression(minPrecedence int) (Object, error) {
	expression := invalid
	var err error

	if p.isUnary(p.tk) {
		unaryOp := p.tk
		if err := p.next(); err != nil {
			return invalid, err
		}
		argument, err := p.subExpression(10)
		if err != nil {
			return invalid, err
		} else if argument.Kind == Invalid {
			return invalid, p.expectedErr("expression")
		}
		expression = Object{Kind: Unary, Name: unaryOp.String(), Value: &argument, Pos: unaryOp.loc}
	}

	if expression.Kind == Invalid {
		expression, err = p.primaryExpression()
		if err != nil {
			return invalid, err
		}
		if expression.Kind == Invalid {
			expression, err = p.prefixExpression()
			if err != nil {
				return invalid, err
			}
		}
	}

	// This is not a valid left hand expression.
	if expression.Kind == Invalid {
		return invalid, nil
	}

	if p.tk.t == tkSpread {
		pos := p.tk.loc
		if err := p.next(); err != nil {
			return invalid, err
		}
		return Object{Kind: Spread, Value: &expression, Pos: pos}, nil
	} else if p.tk.t == '?' {
		p.pushLoc()
		if err := p.next(); err != nil {
			return invalid, err
		}
		first, err := p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		if err = p.expect(':'); err != nil {
			return invalid, err
		}
		second, err := p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		return Object{Kind: Ternary, Vals: []Object{expression, first, second}, Pos: p.popLoc()}, nil
	}

	for {
		operator := p.tk
		precedence := binaryPrecedence[operator.String()]
		if precedence == 0 || precedence <= minPrecedence {
			break
		}
		if operator.String() == "^" { // Right-hand precedence operators
			precedence--
		}
		if err := p.next(); err != nil {
			return invalid, err
		}
		right, err := p.subExpression(precedence)
		if err != nil {
			return invalid, err
		} else if right.Kind == Invalid {
			return invalid, p.expectedErr("expression")
		}
		expression = Object{
			Kind: Binary,
			Name: operator.String(),
			Vals: []Object{expression, right},
			Pos:  operator.loc,
		}
	}
	return expression, nil
}

func (p *parser) primaryExpression() (Object, error) {
	if p.tk.t == tkString {
		val := Object{Kind: String, StringValue: p.tk.stringValue, Pos: p.tk.loc}
		return val, p.next()
	} else if p.tk.t == tkNil {
		return Object{Kind: Nil, Pos: p.tk.loc}, p.next()
	} else if p.tk.t == tkNumber {
		val := Object{Kind: Number, NumberValue: p.tk.numberValue, Pos: p.tk.loc}
		return val, p.next()
	} else if p.tk.t == tkTrue || p.tk.t == tkFalse {
		val := Object{Kind: Bool, BoolValue: p.tk.t == tkTrue, Pos: p.tk.loc}
		return val, p.next()
	} else if p.tk.t == tkFunction {
		return p.functionDeclaration()
	} else if p.tk.t == '{' {
		return p.tableConstructor()
	}
	return invalid, nil
}

func (p *parser) prefixExpression() (Object, error) {
	var base Object
	var err error

	if p.tk.t == tkName {
		base, err = p.identifier()
		if err != nil {
			return invalid, err
		}
	} else if p.tk.t == '(' {
		if err := p.next(); err != nil {
			return invalid, err
		}
		base, err = p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		if err = p.expect(')'); err != nil {
			return invalid, err
		}
	} else {
		return invalid, nil
	}

	for {
		newBase, err := p.prefixExpressionPart(base)
		if err != nil {
			return invalid, err
		} else if newBase.Kind == Invalid {
			break
		}
		base = newBase
	}

	return base, nil
}

func (p *parser) prefixExpressionPart(base Object) (Object, error) {
	switch p.tk.t {
	case '[':
		p.pushLoc()
		if err := p.next(); err != nil {
			return invalid, err
		}
		expression, err := p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		if p.tk.t == ':' {
			p.pushLoc()
			if err := p.next(); err != nil {
				return invalid, err
			}
			rng, err := p.expectedExpression()
			if err != nil {
				return invalid, err
			}
			expression = Object{Kind: Range, Vals: []Object{expression, rng}, Pos: p.popLoc()}
		}
		err = p.expect(']')
		return Object{Kind: Index, Vals: []Object{base, expression}, Pos: p.popLoc()}, err
	case '.':
		p.pushLoc()
		if err := p.next(); err != nil {
			return invalid, err
		}
		identifier, err := p.identifier()
		return Object{Kind: Member, Vals: []Object{base, identifier}, Pos: p.popLoc()}, err
	case '(': // args
		return p.callExpression(base)
	default:
		return invalid, nil
	}
}

func (p *parser) callExpression(base Object) (Object, error) {
	startLoc := p.prev.loc
	if p.tk.loc[0] != p.prev.loc[0] {
		return invalid, p.parseError("ambiguous syntax for function call. Put the open paren on the same line as your function name")
	}
	if err := p.next(); err != nil {
		return invalid, err
	}

	args := []Object{}
	expression, err := p.expression()
	if err != nil {
		return invalid, err
	}
	if expression.Kind != Invalid {
		args = append(args, expression)
	}
	for p.tk.t == ',' {
		if err := p.next(); err != nil {
			return invalid, err
		}
		expression, err = p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		args = append(args, expression)
	}

	p.expect(')')
	return Object{Kind: FuncCall, Value: &base, Vals: args, Pos: p.endLoc(startLoc)}, nil
}

func (p *parser) tableConstructor() (Object, error) {
	if err := p.nextIf('{'); err != nil {
		return invalid, err
	}

	loc := p.prev.loc
	table := Object{Kind: Table}
	for {
		if p.tk.t == '[' {
			p.pushLoc()
			if err := p.next(); err != nil {
				return invalid, err
			}
			key, err := p.expectedExpression()
			if err != nil {
				return invalid, err
			}
			if err = p.expect(']'); err != nil {
				return invalid, err
			}
			if err = p.expect(':'); err != nil {
				return invalid, err
			}
			value, err := p.expectedExpression()
			table.Vals = append(table.Vals, Object{Kind: TableKey, Key: &key, Value: &value, Pos: p.popLoc()})
		} else if p.tk.t == tkName {
			p.pushLoc()
			tk, err := p.lookAhead()
			if err != nil {
				return invalid, err
			}
			if tk.t == ':' {
				key, err := p.identifier()
				if err != nil {
					return invalid, err
				}
				if err := p.next(); err != nil {
					return invalid, err
				}
				value, err := p.expectedExpression()
				if err != nil {
					return invalid, err
				}
				table.Vals = append(table.Vals, Object{Kind: TableKey, Key: &key, Value: &value, Pos: p.popLoc()})
			} else {
				value, err := p.expectedExpression()
				if err != nil {
					return invalid, err
				}
				table.Vals = append(table.Vals, Object{Kind: TableValue, Value: &value, Pos: p.popLoc()})
			}
		} else {
			p.pushLoc()
			value, err := p.expression()
			if err != nil {
				return invalid, err
			}
			if value.Kind == Invalid {
				break
			}
			table.Vals = append(table.Vals, Object{Kind: TableValue, Value: &value, Pos: p.popLoc()})
		}
		if p.tk.t == ',' || p.tk.t == ';' {
			if err := p.next(); err != nil {
				return invalid, err
			}
			continue
		}
		break
	}
	table.Pos = p.endLoc(loc)
	return table, p.expect('}')
}

func (p *parser) functionDeclaration() (Object, error) {
	var name Object
	var err error
	p.pushLoc()

	if err := p.nextIf(tkFunction); err != nil {
		return invalid, err
	}
	if p.tk.t != '(' {
		name, err = p.functionName()
		if err != nil {
			return invalid, err
		}
	}

	if err := p.expect('('); err != nil {
		return invalid, err
	}

	parameters := []Object{}
	if p.tk.t != ')' {
		for {
			if p.tk.t == tkName {
				parameter, err := p.identifier()
				if err != nil {
					return invalid, err
				}
				if p.tk.t == tkSpread {
					if err := p.nextIf(tkSpread); err != nil {
						return invalid, err
					}
					parameter.Name += "..."
					parameters = append(parameters, parameter)
					if err := p.expect(')'); err != nil {
						return invalid, err
					}
					break
				}

				parameters = append(parameters, parameter)

				if p.tk.t == ',' {
					if err := p.next(); err != nil {
						return invalid, err
					}
					continue
				}
			} else {
				return invalid, p.expectedErr("<name> or '...'")
			}
			if err := p.expect(')'); err != nil {
				return invalid, err
			}
			break
		}
	} else if err := p.next(); err != nil {
		return invalid, err
	}

	body, catches, err := p.block()
	if err != nil {
		return invalid, err
	}
	if err := p.expect(tkEnd); err != nil {
		return invalid, err
	}

	return Object{
		Kind:    FuncDef,
		Value:   &name,
		Vars:    parameters,
		Block:   body,
		Catches: catches,
		Pos:     p.popLoc(),
	}, nil
}

func (p *parser) functionName() (Object, error) {
	var base Object
	p.pushLoc()
	base, err := p.identifier()
	if err != nil {
		return invalid, err
	}

	for p.tk.t == '.' {
		if err := p.next(); err != nil {
			return invalid, err
		}
		name, err := p.identifier()
		if err != nil {
			return invalid, err
		}
		base = Object{Kind: Member, Vals: []Object{base, name}, Pos: p.popLoc()}
	}
	return base, nil
}

func (p *parser) returnStatement() (Object, error) {
	var expressions = []Object{}
	p.pushLoc()
	if err := p.nextIf(tkReturn); err != nil {
		return invalid, err
	}
	if p.tk.t != tkEnd {
		expression, err := p.expression()
		if err != nil {
			return invalid, err
		}
		if expression.Kind != Invalid {
			expressions = append(expressions, expression)
		}
		for p.tk.t == ',' {
			if err := p.next(); err != nil {
				return invalid, err
			}
			expression, err = p.expectedExpression()
			if err != nil {
				return invalid, err
			}
			expressions = append(expressions, expression)
		}
		if err := p.nextIf(';'); err != nil {
			return invalid, err
		}
	}
	return Object{Kind: Return, Vals: expressions, Pos: p.popLoc()}, nil
}

func (p *parser) whileStatement() (Object, error) {
	p.pushLoc()
	if err := p.nextIf(tkWhile); err != nil {
		return invalid, err
	}
	p.inLoop++
	defer func() { p.inLoop-- }()
	if condition, err := p.expectedExpression(); err != nil {
		return invalid, err
	} else if err = p.expect(tkDo); err != nil {
		return invalid, err
	} else if body, catches, err := p.block(); err != nil {
		return invalid, err
	} else if err = p.expect(tkEnd); err != nil {
		return invalid, err
	} else {
		return Object{Kind: While, Cond: &condition, Block: body, Catches: catches, Pos: p.popLoc()}, nil
	}
}

func (p *parser) ifStatement() (Object, error) {
	p.pushLoc()
	p.pushLoc()
	if err := p.nextIf(tkIf); err != nil {
		return invalid, err
	}
	statement := Object{Kind: If}

	condition, err := p.expectedExpression()
	if err != nil {
		return invalid, err
	}
	err = p.expect(tkThen)
	if err != nil {
		return invalid, err
	}
	body, catches, err := p.block()
	if err != nil {
		return invalid, err
	}
	statement.Block = append(statement.Block, Object{
		Kind:    IfClause,
		Cond:    &condition,
		Block:   body,
		Catches: catches,
		Pos:     p.popLoc(),
	})

	for p.tk.t == tkElseif {
		p.pushLoc()
		if err := p.next(); err != nil {
			return statement, err
		}
		condition, err := p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		err = p.expect(tkThen)
		if err != nil {
			return invalid, err
		}
		body, catches, err = p.block()
		if err != nil {
			return invalid, err
		}
		statement.Block = append(statement.Block, Object{
			Kind:    IfClause,
			Cond:    &condition,
			Block:   body,
			Catches: catches,
			Pos:     p.popLoc(),
		})
	}

	if p.tk.t == tkElse {
		p.pushLoc()
		if err := p.next(); err != nil {
			return invalid, err
		}
		body, catches, err = p.block()
		if err != nil {
			return invalid, err
		}
		statement.Block = append(statement.Block, Object{
			Kind:    IfClause,
			Block:   body,
			Catches: catches,
			Pos:     p.popLoc(),
		})
	}

	err = p.expect(tkEnd)
	if err != nil {
		return invalid, err
	}
	statement.Pos = p.popLoc()
	return statement, nil
}

func (p *parser) forStatement() (Object, error) {
	p.pushLoc()
	if err := p.nextIf(tkFor); err != nil {
		return invalid, err
	}
	p.inLoop++
	defer func() { p.inLoop-- }()
	variable, err := p.identifier()
	if err != nil {
		return invalid, err
	}

	if p.tk.t == '=' {
		if err := p.next(); err != nil {
			return invalid, err
		}
		start, err := p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		p.expect(',')
		cond, err := p.expectedExpression()
		if err != nil {
			return invalid, err
		}
		p.expect(',')
		step, err := p.assignmentOrCallStatement()
		if err != nil {
			return invalid, err
		}

		err = p.expect(tkDo)
		if err != nil {
			return invalid, err
		}
		body, catches, err := p.block()
		if err != nil {
			return invalid, err
		}
		err = p.expect(tkEnd)
		if err != nil {
			return invalid, err
		}

		return Object{
			Kind:    ForNum,
			Name:    variable.Name,
			Value:   &start,
			Step:    &step,
			Cond:    &cond,
			Block:   body,
			Catches: catches,
			Pos:     p.popLoc(),
		}, nil
	}

	variables := []Object{variable}
	if p.tk.t == ',' {
		if err := p.next(); err != nil {
			return invalid, err
		}
		v, err := p.identifier()
		if err != nil {
			return invalid, err
		}
		variables = append(variables, v)
	}
	p.expect(tkIn)

	iterator, err := p.expectedExpression()
	if err != nil {
		return invalid, err
	}

	err = p.expect(tkDo)
	if err != nil {
		return invalid, err
	}
	body, catches, err := p.block()
	if err != nil {
		return invalid, err
	}
	err = p.expect(tkEnd)
	if err != nil {
		return invalid, err
	}

	return Object{
		Kind:    ForIn,
		Vars:    variables,
		Value:   &iterator,
		Block:   body,
		Catches: catches,
		Pos:     p.popLoc(),
	}, nil
}

func (p *parser) doStatement() (Object, error) {
	p.pushLoc()
	if err := p.nextIf(tkDo); err != nil {
		return invalid, err
	} else if body, catches, err := p.block(); err != nil {
		return invalid, err
	} else if err := p.expect(tkEnd); err != nil {
		return invalid, err
	} else {
		return Object{Kind: Do, Block: body, Catches: catches, Pos: p.popLoc()}, nil
	}
}

func (p *parser) consumecleanupStatement() ([]Object, error) {
	catches := []Object{}
	for p.tk.t == tkCleanup {
		catch, err := p.cleanupStatement()
		if err != nil {
			return []Object{}, err
		}
		catches = append(catches, catch)
	}
	return catches, nil
}

func (p *parser) cleanupStatement() (Object, error) {
	p.pushLoc()
	if err := p.nextIf(tkCleanup); err != nil {
		return invalid, err
	}

	var name string
	variable, err := p.identifier()
	if err != nil {
		return invalid, err
	}

	if p.tk.t == '=' {
		name = variable.Name
		if err := p.next(); err != nil {
			return invalid, err
		}
		variable, err = p.identifier()
		if err != nil {
			return invalid, err
		}
	}

	errorClasses := []Object{variable}
	for p.tk.t == ',' {
		if err := p.next(); err != nil {
			return invalid, err
		}
		variable, err = p.identifier()
		if err != nil {
			return invalid, err
		}
		errorClasses = append(errorClasses, variable)
	}

	if err := p.expect(tkDo); err != nil {
		return invalid, err
	}

	body, catches, err := p.block()
	if err != nil {
		return invalid, err
	}

	return Object{Kind: Cleanup, Name: name, Block: body, Catches: catches, Vars: errorClasses, Pos: p.popLoc()}, nil
}

func (p *parser) isUnary(tk token) bool {
	switch tk.t {
	case '!', '#', '-', '~', '@':
		return true
	default:
		return false
	}
}

func (p *parser) isAssignShortcut(tk token) bool {
	switch tk.t {
	case tkDecrEq, tkDecrement, tkIncrEq, tkIncrement, tkShiftLeft:
		return true
	default:
		return false
	}
}
