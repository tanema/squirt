package runtime

import (
	"fmt"
	"strings"

	"github.com/tanema/squirt/src/lang"
)

type (
	Value  interface{}
	Break  struct{}
	Next   struct{}
	Return struct{ Vals []Value }
	Spread struct{ Table *Table }
	Range  struct{ Start, End int }

	// CVal Allows classes (and maybe funcs) to be uses as values in the same way
	CVal interface {
		Self() CVal
		Super(*Scope, CVal, Value, []Value) Value
		IsA(string) bool
		Type() string
		OpIndex(*Scope, Value) (Value, error)
		OpAssignIndex(*Scope, Value, Value) (Value, error)
		Op(string, *Scope, ...Value) (Value, error)
		ToString(*Scope) string
		ToBoolean(*Scope) bool
	}

	Member struct {
		s      *Scope
		r      *Runtime
		obj    lang.Object
		source Value
		key    Value
	}
)

type Runtime struct {
	isFile   bool
	filepath string
	trace    []string
}

func EvalFile(scope *Scope, filename string) (Value, error) {
	ast, err := lang.ParseFile(filename)
	if err != nil {
		return nil, err
	}
	r := Runtime{filepath: filename, isFile: true}
	r.pushStack("<main>", 0)
	defer r.popStack()
	val, err := r.evalBlock(scope, ast.Block, []lang.Object{})
	if err != nil {
		return nil, err
	}
	if ret, ok := val.(Return); ok {
		if len(ret.Vals) == 1 {
			val = ret.Vals[0]
		} else if len(ret.Vals) > 1 {
			val, _ = ToValue(scope, ret.Vals)
		} else {
			val = nil
		}
	}
	return val, nil

}

func Eval(scope *Scope, in string) (Value, error) {
	ast, err := lang.ParseStr(in)
	if err != nil {
		return nil, err
	}
	r := Runtime{filepath: "<input>"}
	r.pushStack("<main>", -1)
	defer r.popStack()
	val, err := r.eval(scope, ast.Block[0])
	if err != nil {
		return nil, err
	}
	if ret, ok := val.(Return); ok {
		if len(ret.Vals) == 1 {
			val = ret.Vals[0]
		} else if len(ret.Vals) > 1 {
			val, _ = ToValue(scope, ret.Vals)
		} else {
			val = nil
		}
	}
	return val, nil
}

func (r *Runtime) pushStack(name string, lineno int) {
	r.trace = append(r.trace, fmt.Sprintf("%v:%v in %v", r.filepath, lineno, name))
}

func (r *Runtime) popStack() {
	r.trace = r.trace[:len(r.trace)-1]
}

func (r *Runtime) runtimeError(scope *Scope, obj lang.Object, msg string, data ...interface{}) error {
	msg = fmt.Sprintf(msg, data...)
	inst, _ := create(scope, "RuntimeError", msg)
	return RuntimeErr{
		isFile:     r.isFile,
		file:       r.filepath,
		source:     obj,
		errorClass: "RuntimeError",
		msg:        msg,
		errInst:    inst,
		stacktrace: r.trace,
	}
}

func (r *Runtime) wrapErr(scope *Scope, obj lang.Object, err error) error {
	if err == nil {
		return nil
	} else if _, isRuntime := err.(RuntimeErr); isRuntime {
		return err
	} else if inst, isinst := err.(*Instance); isinst && inst.IsA("Error") {
		msg, _ := inst.data["message"].(string)
		return RuntimeErr{
			isFile:     r.isFile,
			file:       r.filepath,
			source:     obj,
			errorClass: inst.class.name,
			errInst:    inst,
			msg:        msg,
			stacktrace: r.trace,
		}
	}
	return r.runtimeError(scope, obj, err.Error())
}

func (r *Runtime) evalBlock(scope *Scope, block, catches []lang.Object) (Value, error) {
	for _, obj := range block {
		result, err := r.eval(scope, obj)
		if err != nil {
			if userErr, isRuntime := err.(RuntimeErr); isRuntime {
				for _, catch := range catches {
					for _, errClass := range catch.Vars {
						if userErr.errInst.IsA(errClass.Name) {
							if catch.Name != "" {
								scope = scope.Child(map[string]Value{catch.Name: userErr.errInst})
							}
							return r.evalBlock(scope, catch.Block, catch.Catches)
						}
					}
				}
			}
			return nil, err
		}
		switch result.(type) {
		case Break, Return, Next:
			return result, nil
		}
	}
	return nil, nil
}

func (r *Runtime) eval(scope *Scope, object lang.Object) (result Value, err error) {
	switch object.Kind {
	case lang.Assignment:
		return nil, r.evalAssign(scope, object)
	case lang.FuncCall:
		return r.evalFuncCall(scope, object)
	case lang.FuncDef:
		return r.evalFuncDef(scope, object)
	case lang.If:
		return r.evalIfStatement(scope, object)
	case lang.Do:
		return r.evalBlock(scope, object.Block, object.Catches)
	case lang.ForIn:
		return r.evalForIn(scope, object)
	case lang.ForNum:
		return r.evalForNum(scope, object)
	case lang.While:
		return r.evalWhile(scope, object)
	case lang.Binary:
		return r.evalOperator(scope, object, &object.Vals[0], &object.Vals[1])
	case lang.Unary:
		return r.evalOperator(scope, object, object.Value, nil)
	case lang.Table:
		return r.evalTableStatement(scope, object)
	case lang.Index:
		return r.evalIndexStatement(scope, object.Vals[0], object.Vals[1], false)
	case lang.Member:
		return r.evalIndexStatement(scope, object.Vals[0], object.Vals[1], true)
	case lang.Return:
		return r.evalReturnStatement(scope, object)
	case lang.Identifier:
		return scope.Get(object.Name), nil
	case lang.String:
		return r.evalStringLit(scope, object.StringValue)
	case lang.Bool:
		return ToValue(scope, object.BoolValue)
	case lang.Break:
		return Break{}, nil
	case lang.Next:
		return Next{}, nil
	case lang.Number:
		return ToValue(scope, object.NumberValue)
	case lang.Nil:
		return ToValue(scope, nil)
	case lang.Spread:
		return r.evalSpreadStatement(scope, object)
	case lang.Range:
		return r.evalRange(scope, object)
	case lang.ClassDef:
		return r.evalClassDef(scope, object)
	case lang.Ternary:
		return r.evalTernaryStatement(scope, object)
	default:
		return nil, r.runtimeError(scope, object, "missed object kind %v, this means squirt is broken and it is not your code", object.Kind)
	}
}

func (r *Runtime) evalStringLit(scope *Scope, str string) (Value, error) {
	str, err := lang.Interpolate(str, func(s string) (string, error) {
		ast, err := lang.ParseStr(s)
		if err != nil {
			return "", err
		}
		expr, err := r.eval(scope, ast.Block[0])
		if err != nil {
			return "", err
		}
		return toString(scope, expr), nil
	})
	if err != nil {
		return nil, err
	}
	return create(scope, "String", str)
}

func (r *Runtime) evalAssign(scope *Scope, assign lang.Object) error {
	varInx := 0
	roundup := []Value{}

	for i, v := range assign.Vals {
		val, err := r.eval(scope, v)
		if err != nil {
			return err
		}
		vals := []Value{val}
		if res, ok := val.(Spread); ok {
			vals = append(res.Table.Arr)
		}

		for j, val := range vals {
			valsLeft := (len(assign.Vals) - i - 1) + (len(vals) - j)
			if varInx == len(assign.Vars)-1 && (len(roundup) > 0 || valsLeft > 1) {
				roundup = append(roundup, val)
				if j == len(vals)-1 && i == len(assign.Vals)-1 {
					val, _ = create(scope, "Table", roundup...)
				} else {
					continue
				}
			}
			target := assign.Vars[varInx]
			switch target.Kind {
			case lang.Identifier:
				scope.Set(target.Name, val)
			case lang.Member, lang.Index:
				inx, err := r.evalIndexStatement(scope, target.Vals[0], target.Vals[1], target.Kind == lang.Member)
				if err != nil {
					return r.wrapErr(scope, assign, err)
				}
				_, err = inx.set(val)
				return r.wrapErr(scope, assign, err)
			default:
				return r.runtimeError(scope, assign, "cannot assign to type %v", target.Kind)
			}
			varInx++
		}
	}
	return nil
}

func (r *Runtime) evalFuncCall(scope *Scope, call lang.Object) (Value, error) {
	fnCall, err := r.eval(scope, *call.Value)
	if err != nil {
		return nil, err
	}
	args := []Value{}
	for _, ex := range call.Vals {
		a, err := r.eval(scope, ex)
		if err != nil {
			return nil, err
		}
		if spr, ok := a.(Spread); ok {
			args = append(args, spr.Table.Arr...)
		} else {
			args = append(args, a)
		}
	}

	var self CVal = nil
	if mem, ok := fnCall.(Member); ok {
		if fnCall, err = mem.get(); err != nil {
			return nil, r.wrapErr(scope, call, err)
		}
		self, _ = mem.source.(CVal)
	}

	if fn, is := fnCall.(*Func); is {
		res, err := fn.call(scope, self, args)
		return res, r.wrapErr(scope, call, err)
	} else if inst, is := fnCall.(*Instance); is {
		res, err := inst.Op("__call", scope, args...)
		return res, r.wrapErr(scope, call, err)
	}
	return nil, r.runtimeError(scope, call, "tried to call a non callable object %v", typeOf(fnCall))
}

func evalFuncDefParams(fnSt lang.Object) (params []string, vararg bool) {
	for i, p := range fnSt.Vars {
		ident := p.Name
		if i == len(fnSt.Vars)-1 && strings.HasSuffix(ident, "...") {
			vararg = true
			ident = strings.TrimRight(ident, "...")
		}
		params = append(params, ident)
	}
	return
}

func (r *Runtime) evalFuncDef(scope *Scope, fnSt lang.Object) (Value, error) {
	paramdefs, vararg := evalFuncDefParams(fnSt)
	var fnName string
	if fnSt.Value != nil {
		fnName = fnSt.Value.Name
	}
	fn := &Func{
		Name:   fnName,
		LineNo: fnSt.Pos[0],
		Params: paramdefs,
		Vararg: vararg,
		Fn: func(s *Scope, self CVal, args []Value) (Value, error) {
			r.pushStack(fnName, fnSt.Pos[0])
			defer r.popStack()
			params, err := mapParams(s, paramdefs, args, vararg)
			if err != nil {
				return nil, err
			}
			if self != nil {
				params["self"] = self.Self()
				params["super"] = self.Super(scope, self, fnName, args)
			}
			return r.evalBlock(scope.Child(params), fnSt.Block, fnSt.Catches)
		},
	}

	if fnSt.Value != nil {
		if fnSt.Value.Kind == lang.Identifier {
			scope.Set(fnSt.Value.Name, fn)
			return fn, nil
		} else if fnSt.Value.Kind == lang.Member {
			mempre, err := r.evalIndexStatement(scope, fnSt.Value.Vals[0], fnSt.Value.Vals[1], true)
			if err != nil {
				return nil, err
			}
			_, err = mempre.set(fn)
			return fn, r.wrapErr(scope, fnSt, err)
		}
	}

	return fn, nil
}

func (r *Runtime) evalIfStatement(scope *Scope, ifSt lang.Object) (Value, error) {
	for i, st := range ifSt.Block {
		if i == len(ifSt.Block)-1 && st.Cond == nil {
			return r.evalBlock(scope, st.Block, st.Catches)
		} else if cond, err := r.eval(scope, *st.Cond); err != nil {
			return nil, err
		} else if !toBool(scope, cond) {
			continue
		}
		return r.evalBlock(scope, st.Block, st.Catches)
	}
	return nil, nil
}

func (r *Runtime) evalTernaryStatement(scope *Scope, tern lang.Object) (Value, error) {
	if cond, err := r.eval(scope, tern.Vals[0]); err != nil {
		return nil, err
	} else if toBool(scope, cond) {
		return r.eval(scope, tern.Vals[1])
	}
	return r.eval(scope, tern.Vals[2])
}

func (r *Runtime) evalForNum(scope *Scope, forNum lang.Object) (Value, error) {
	startVal, err := r.eval(scope, *forNum.Value)
	if err != nil {
		return nil, nil
	}
	scope.Set(forNum.Name, startVal)
	defer scope.Set(forNum.Name, nil)
	for {
		cond, err := r.eval(scope, *forNum.Cond)
		if err != nil {
			return nil, err
		}
		if !toBool(scope, cond) {
			break
		}

		result, err := r.evalBlock(scope, forNum.Block, forNum.Catches)
		if err != nil {
			return nil, err
		}
		switch result.(type) {
		case Break:
			break
		case Return:
			return result, nil
		case Next:
			continue
		}

		if _, err := r.eval(scope, *forNum.Step); err != nil {
			return nil, err
		}
	}

	return nil, nil
}

func (r *Runtime) evalForIn(scope *Scope, forIn lang.Object) (Value, error) {
	data, err := r.eval(scope, *forIn.Value)
	if err != nil {
		return nil, err
	}

	inst, is := data.(*Instance)
	if !is || inst.class.name != "Table" {
		return nil, r.runtimeError(scope, forIn, "used for-in loop on non table data type")
	}
	table := inst.data["_tbl"].(*Table)

	i := 0
	keyVar := forIn.Vars[0].Name
	valVar := forIn.Vars[1].Name
	defer scope.Set(keyVar, nil)
	defer scope.Set(valVar, nil)

	for {
		if i >= len(table.Keys) {
			break
		}

		key := table.Keys[i]
		val := table.Values[i]
		scope.Set(keyVar, key)
		scope.Set(valVar, val)

		result, err := r.evalBlock(scope, forIn.Block, forIn.Catches)
		if err != nil {
			return nil, err
		}
		switch result.(type) {
		case Break:
			break
		case Return:
			return result, nil
		case Next:
			continue
		}
		i++
	}

	return nil, nil
}

func (r *Runtime) evalWhile(scope *Scope, while lang.Object) (Value, error) {
	for {
		if cond, err := r.eval(scope, *while.Cond); err != nil {
			return nil, err
		} else if !toBool(scope, cond) {
			break
		}

		result, err := r.evalBlock(scope, while.Block, while.Catches)
		if err != nil {
			return nil, err
		}
		switch result.(type) {
		case Break:
			break
		case Return:
			return result, nil
		case Next:
			continue
		}
	}
	return nil, nil
}

func (r *Runtime) evalInstanceValue(scope *Scope, val *lang.Object) (Value, bool, error) {
	if val == nil {
		return nil, false, nil
	}

	obj, err := r.eval(scope, *val)
	if err != nil {
		return nil, false, err
	}

	if mem, ok := obj.(Member); ok {
		if obj, err = mem.get(); err != nil {
			return nil, false, r.wrapErr(scope, *val, err)
		}
	}

	final, isInst := obj.(*Instance)
	if !isInst {
		return obj, false, nil
	}
	return final, true, nil
}

func (r *Runtime) evalOperator(scope *Scope, obj lang.Object, leftVal, rightVal *lang.Object) (Value, error) {
	leftObj, isInst, err := r.evalInstanceValue(scope, leftVal)
	if err != nil {
		if obj.Name == "@" {
			if errInst, isInst := err.(*Instance); isInst {
				return errInst, nil
			} else if rerr, isRuntime := err.(RuntimeErr); isRuntime {
				return create(scope, "RuntimeError", rerr.msg)
			}
			return create(scope, "RuntimeError", err.Error())
		}
		return nil, err
	}

	if obj.Name == "@" {
		n, _ := ToValue(scope, nil)
		if spr, ok := leftObj.(Spread); ok {
			spr.Table.Arr = append([]Value{n}, spr.Table.Arr...)
		} else if leftObj == nil {
			leftObj = Spread{Table: &Table{Arr: []Value{n}}}
		} else {
			leftObj = Spread{Table: &Table{Arr: []Value{n, leftObj}}}
		}
		return leftObj, nil
	}

	if !isInst {
		return nil, r.runtimeError(scope, obj, "left operand is an invalid value")
	}
	left := leftObj.(*Instance)

	if obj.Kind == lang.Unary {
		switch obj.Name {
		case "~":
			return left.Op("__bitnot", scope)
		case "!":
			return ToValue(scope, !toBool(scope, left))
		case "#":
			return left.Op("__len", scope)
		default:
			return nil, r.runtimeError(scope, obj, "unsupported unary %v", obj.Name)
		}
	} else if obj.Name == "and" {
		if !toBool(scope, left) {
			return left, nil
		}
		rightObj, _, err := r.evalInstanceValue(scope, rightVal)
		return rightObj, err
	} else if obj.Name == "or" {
		if toBool(scope, left) {
			return left, nil
		}
		rightObj, _, err := r.evalInstanceValue(scope, rightVal)
		return rightObj, err
	}

	rightObj, isInst, err := r.evalInstanceValue(scope, rightVal)
	if err != nil {
		return nil, err
	} else if !isInst {
		return nil, r.runtimeError(scope, obj, "right operand is an invalid value")
	}
	right := rightObj.(*Instance)

	switch obj.Name {
	case "+":
		return left.Op("__add", scope, right)
	case "-":
		return left.Op("__sub", scope, right)
	case "/":
		return left.Op("__div", scope, right)
	case "*":
		return left.Op("__mul", scope, right)
	case "^":
		return left.Op("__exp", scope, right)
	case "%":
		return left.Op("__mod", scope, right)
	case "<<":
		return left.Op("__shiftleft", scope, right)
	case ">>":
		return left.Op("__shiftright", scope, right)
	case "&":
		return left.Op("__and", scope, right)
	case "~":
		return left.Op("__xor", scope, right)
	case "|":
		return left.Op("__or", scope, right)
	case "<", "<=", ">", ">=":
		val, err := left.Op("__compare", scope, right)
		if err != nil {
			return nil, err
		}
		inst, isInst := val.(*Instance)
		if !isInst || !inst.IsA("Number") {
			return ToValue(scope, false)
		}
		cmpr := inst.data["_val"].(float64)
		if (cmpr <= -1 && obj.Name[0] == '<') ||
			(cmpr >= 1 && obj.Name[0] == '>') ||
			(cmpr == 0 && strings.Contains(obj.Name, "=")) {
			return ToValue(scope, true)
		}
		return ToValue(scope, false)
	case "==":
		return left.Op("__eq", scope, right)
	case "!=":
		val, err := left.Op("__eq", scope, right)
		if err != nil {
			return nil, err
		}
		return ToValue(scope, !toBool(scope, val))
	case "!":
		return ToValue(scope, !toBool(scope, left))
	case "#":
		return left.Op("__len", scope)
	}
	return nil, r.runtimeError(scope, obj, "undefined operation %v %v %v", typeOf(left), obj.Name, typeOf(right))
}

func (r *Runtime) evalTable(scope *Scope, tbl lang.Object) (*Table, error) {
	var err error
	table := &Table{}
	for _, val := range tbl.Vals {
		switch val.Kind {
		case lang.TableValue:
			value, err := r.eval(scope, *val.Value)
			if err != nil {
				return nil, err
			}
			if spr, ok := value.(Spread); ok {
				table.add(scope, spr.Table)
			} else {
				table.Arr = append(table.Arr, value)
			}
		case lang.TableKey:
			var key Value
			var value Value
			if val.Key.Kind == lang.Identifier {
				key, _ = create(scope, "String", val.Key.Name)
			} else {
				key, err = r.eval(scope, *val.Key)
				if err != nil {
					return nil, err
				}
			}
			value, err = r.eval(scope, *val.Value)
			if err != nil {
				return nil, err
			}
			table.assign(scope, key, value)
		}
	}
	return table, nil
}

func (r *Runtime) evalTableStatement(scope *Scope, tbl lang.Object) (*Instance, error) {
	table, err := r.evalTable(scope, tbl)
	if err != nil {
		return nil, err
	}
	return create(scope, "Table", table)
}

func (r *Runtime) evalIndexStatement(scope *Scope, base, index lang.Object, isMember bool) (Member, error) {
	indexable, err := r.eval(scope, base)
	if err != nil {
		return Member{}, r.wrapErr(scope, base, err)
	}
	if mem, isMember := indexable.(Member); isMember {
		if indexable, err = mem.get(); err != nil {
			return Member{}, r.wrapErr(scope, base, err)
		}
	}
	var key Value
	if isMember && index.Kind == lang.Identifier {
		key, _ = create(scope, "String", index.Name)
	} else {
		key, err = r.eval(scope, index)
		if err != nil {
			return Member{}, r.wrapErr(scope, index, err)
		}
	}

	if indexable == nil {
		return Member{}, r.runtimeError(scope, index, "cannot index nil")
	}

	return Member{s: scope, r: r, obj: index, source: indexable, key: key}, nil
}

func (r *Runtime) evalReturnStatement(scope *Scope, ret lang.Object) (Return, error) {
	result := Return{}
	for _, v := range ret.Vals {
		val, err := r.eval(scope, v)
		if err != nil {
			return result, err
		}
		result.Vals = append(result.Vals, val)
	}
	return result, nil
}

func (r *Runtime) evalSpreadStatement(scope *Scope, spread lang.Object) (Spread, error) {
	val, err := r.eval(scope, *spread.Value)
	if err != nil {
		return Spread{}, err
	} else if tbl, ok := val.(*Table); !ok {
		return Spread{}, r.runtimeError(scope, spread, "spread operator used on non table value")
	} else {
		return Spread{Table: tbl}, nil
	}
}
func (r *Runtime) evalRange(scope *Scope, obj lang.Object) (Range, error) {
	startVal, err := r.eval(scope, obj.Vals[0])
	if err != nil {
		return Range{}, err
	}
	start, isStartInt := isIntKey(startVal)
	if !isStartInt {
		return Range{}, r.runtimeError(scope, obj, "start in range is not a non decimal number")
	}
	endVal, err := r.eval(scope, obj.Vals[1])
	if err != nil {
		return Range{}, err
	}
	end, isEndInt := isIntKey(endVal)
	if !isEndInt {
		return Range{}, r.runtimeError(scope, obj, "end in range is not a non decimal number")
	}
	if end < start {
		return Range{}, r.runtimeError(scope, obj, "range can only be positive")
	}
	return Range{Start: start, End: end}, nil
}

func (r *Runtime) evalClassDef(scope *Scope, classdef lang.Object) (*Class, error) {
	var parent *Class
	if classdef.Parent != "" {
		var err error
		parent, err = findClass(scope, classdef.Parent)
		if err != nil {
			return nil, err
		}
	}

	attrs := []*Attribute{}
	for _, obj := range classdef.Block {
		switch obj.Kind {
		case lang.FuncDef:
			paramdefs, vararg := evalFuncDefParams(obj)
			name := obj.Value.Name
			lineno := obj.Pos[0]
			block := obj.Block
			catches := obj.Catches
			attrs = append(attrs, Attr(obj.Value.Name, &Func{
				ClassName: classdef.Name,
				Name:      name,
				LineNo:    lineno,
				Params:    paramdefs,
				Vararg:    vararg,
				Fn: func(s *Scope, self CVal, args []Value) (Value, error) {
					r.pushStack(classdef.Name+"."+name, lineno)
					defer r.popStack()
					params, err := mapParams(s, paramdefs, args, vararg)
					if err != nil {
						return nil, err
					}
					params["self"] = self.Self()
					params["super"] = self.Super(scope, self, name, args)
					return r.evalBlock(scope.Child(params), block, catches)
				},
			}, &Refinement{constant: true}))
		case lang.AttrDef:
			var val Value
			if obj.Value != nil {
				v, err := r.eval(scope, *obj.Value)
				if err != nil {
					return nil, err
				}
				val = v
			}
			refine, err := parseRefinement(scope, r, obj.Cond)
			if err != nil {
				return nil, err
			}
			attrs = append(attrs, Attr(obj.Name, val, refine))
		default:
			return nil, r.runtimeError(scope, obj, "unexpected object kind %v in class declr, this means squirt is broken and it is not your code", obj.Kind)
		}
	}

	class := CreateClass(classdef.Name, parent, attrs...)
	scope.Set(classdef.Name, class)
	return class, nil
}

func (mem *Member) get() (Value, error) {
	if iface, ok := mem.source.(CVal); ok {
		return iface.OpIndex(mem.s, mem.key)
	}
	return nil, mem.r.runtimeError(mem.s, mem.obj, "cannot index %v on %v", toString(mem.s, mem.key), typeOf(mem.source))
}

func (mem *Member) set(val Value) (Value, error) {
	if iface, ok := mem.source.(CVal); ok {
		return iface.OpAssignIndex(mem.s, mem.key, val)
	}
	return nil, mem.r.runtimeError(mem.s, mem.obj, "cannot assign index %v on %v", toString(mem.s, mem.key), typeOf(mem.source))
}
