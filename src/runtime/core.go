package runtime

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/tanema/squirt/src/lang"
)

var fns = map[string]FnSig{
	"new":      stdNew,
	"spill":    stdSpill,
	"eval":     stdEval,
	"require":  stdRequire,
	"print":    stdPrint,
	"typeof":   stdTypeOf,
	"delete":   stdDelete,
	"tostring": stdToString,
	"tonumber": stdParseNum,
}

var (
	ArgumentError     = CreateClass("ArgumentError", ErrorClass)
	RuntimeErrorClass = CreateClass("RuntimeError", ErrorClass)
)

// DefaultNamespace generate an evironment with the core function and variable declarations defined
func DefaultNamespace(out io.StringWriter) *Scope {
	if out == nil {
		out = os.Stdout
	}
	def := newScope(nil, nil, out)
	for method, fn := range fns {
		def.Set(method, Fn(method, fn))
	}
	def.Set("Boolean", BooleanClass)
	def.Set("Nil", NilClass)
	def.Set("Number", NumberClass)
	def.Set("String", StringClass)
	def.Set("Table", TableClass)

	def.Set("Error", ErrorClass)
	def.Set("ArgumentError", ArgumentError)
	def.Set("RuntimeError", RuntimeErrorClass)
	return def
}

func stdNew(s *Scope, self CVal, args []Value) (Value, error) {
	if len(args) == 0 {
		return nil, fmt.Errorf("not enough arguments to new")
	}
	cls, ok := args[0].(*Class)
	if !ok {
		return nil, fmt.Errorf("wrong value type %v passed to new", typeOf(args[0]))
	}
	return cls.New(s, args[1:]...)
}

func stdSpill(s *Scope, self CVal, args []Value) (Value, error) {
	if len(args) == 0 {
		return createErr(s, ArgumentError, "not enough arguments to spill")
	} else if inst, ok := args[0].(*Instance); len(args) == 1 && ok && inst.IsA("String") {
		return nil, fmt.Errorf(inst.data["_val"].(string))
	} else if cls, ok := args[0].(*Class); ok {
		inst, err := cls.New(s, args[1:]...)
		if err != nil {
			return nil, err
		}
		if !inst.IsA("Error") {
			return createErr(s, ArgumentError, "cannot spill non-error classes")
		}
		return nil, inst
	}
	return createErr(s, ArgumentError, "bad params passed to spill")
}

func stdEval(s *Scope, self CVal, a []Value) (Value, error) {
	if len(a) == 0 {
		return createErr(s, ArgumentError, "not enough arguments to eval")
	} else if inst, ok := a[0].(*Instance); !ok || !inst.IsA("String") {
		return nil, fmt.Errorf("wrong value type passed to eval")
	} else {
		return Eval(s, inst.data["_val"].(string))
	}
}

func stdPrint(s *Scope, self CVal, a []Value) (Value, error) {
	strList := make([]string, len(a))
	for i, e := range a {
		strList[i] = toString(s, e)
	}
	s.out.WriteString(strings.Join(strList, " ") + "\n")
	return nil, nil
}

func stdDelete(s *Scope, self CVal, a []Value) (Value, error) {
	if len(a) < 2 {
		return nil, nil
	}
	deletable, isit := a[0].(CVal)
	if !isit {
		return nil, nil
	}
	for _, mem := range a[1:] {
		if _, err := deletable.Op("__del", s, mem); err != nil {
			return nil, err
		}
	}
	return nil, nil
}

func stdTypeOf(s *Scope, self CVal, a []Value) (Value, error) {
	if len(a) == 0 {
		return nil, fmt.Errorf("not enough arguments to typeof")
	}
	return typeOf(a[0]), nil
}

func typeOf(val Value) string {
	for {
		switch obj := val.(type) {
		case nil:
			return "nil"
		case Member:
			val, _ = obj.get()
		case Spread:
			return "spread"
		case lang.Object:
			return "LANG_OBJECT: " + string(obj.Kind)
		case CVal:
			return obj.Type()
		default:
			return reflect.TypeOf(val).Name()
		}
	}
}

func stdToString(s *Scope, self CVal, a []Value) (Value, error) {
	return toString(s, a[0]), nil
}

func Print(s *Scope, val Value) string {
	return toString(s, val)
}

func toString(s *Scope, val Value) string {
	for {
		switch obj := val.(type) {
		case string:
			return obj
		case Member:
			val, _ = obj.get()
		case CVal:
			return obj.ToString(s)
		case error:
			return obj.Error()
		default:
			return ""
		}
	}
}

func stdParseNum(s *Scope, self CVal, a []Value) (Value, error) {
	return create(s, "Number", a...)
}

func toBool(s *Scope, value Value) bool {
	for {
		switch val := value.(type) {
		case bool:
			return val
		case Member:
			value, _ = val.get()
		case CVal:
			return val.ToBoolean(s)
		default:
			return false
		}
	}
}

func toNumber(obj interface{}) float64 {
	for {
		switch val := obj.(type) {
		case string:
			v, _ := strconv.ParseFloat(val, 64)
			return v
		case float32:
			return float64(val)
		case float64:
			return val
		case int:
			return float64(val)
		case int32:
			return float64(val)
		case int64:
			return float64(val)
		case Member:
			obj, _ = val.get()
		case *Instance:
			if val.class.name == "Number" {
				return val.data["_val"].(float64)
			} else if val.class.name == "String" {
				obj = val.data["_val"].(string)
			} else {
				return 0
			}
		default:
			return 0
		}
	}
}

func ToValue(s *Scope, obj interface{}) (Value, error) {
	switch val := obj.(type) {
	case bool:
		return create(s, "Boolean", val)
	case string:
		return create(s, "String", val)
	case nil:
		return create(s, "Nil")
	case int64, int32, int, float64, float32:
		return create(s, "Number", val)
	case *Table:
		return create(s, "Table", val)
	case []Value:
		return create(s, "Table", val...)
	case map[string]Value:
		return create(s, "Table", val)
	default:
		return obj, nil
	}
}
