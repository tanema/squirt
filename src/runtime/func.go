package runtime

import (
	"fmt"
	"strings"
)

type (
	FnSig func(*Scope, CVal, []Value) (Value, error)
	Func  struct {
		ClassName string
		Name      string
		LineNo    int
		Params    []string
		Vararg    bool
		Std       bool
		Fn        FnSig
	}
)

func (fn *Func) call(s *Scope, self CVal, args []Value) (Value, error) {
	val, err := fn.Fn(s, self, args)
	if err != nil {
		return nil, err
	} else if ret, isRet := val.(Return); isRet && len(ret.Vals) == 1 {
		return ret.Vals[0], nil
	} else if isRet && len(ret.Vals) > 1 {
		return Spread{Table: &Table{Arr: ret.Vals}}, nil
	} else if val != nil && fn.Std {
		return ToValue(s, val)
	}
	return nil, nil
}

func (fn *Func) Type() string                             { return "Func" }
func (fn *Func) IsA(other string) bool                    { return other == "Func" }
func (fn *Func) Self() CVal                               { return fn }
func (fn *Func) Super(*Scope, CVal, Value, []Value) Value { return nil }
func (fn *Func) ToBoolean(s *Scope) bool                  { return true }

func (fn *Func) Op(method string, s *Scope, args ...Value) (Value, error) {
	return nil, fmt.Errorf("cannot use %v on Func", method)
}

func (fn *Func) OpIndex(scope *Scope, key Value) (Value, error) {
	return nil, fmt.Errorf("cannot index on Func")
}
func (fn *Func) OpAssignIndex(scope *Scope, key, val Value) (Value, error) {
	return nil, fmt.Errorf("cannot assign index on Func")
}

func (fn *Func) ToString(s *Scope) string {
	name := fn.Name
	if fn.ClassName != "" {
		name = fn.ClassName + "." + name
	}
	params := []string{}
	for _, p := range fn.Params {
		params = append(params, p)
	}
	if fn.Vararg {
		params[len(params)-1] += "..."
	}
	strParams := "(" + strings.Join(params, ", ") + ")"
	builtin := ""
	if fn.Std {
		builtin = " builtin"
	}
	return "#<func " + name + strParams + builtin + ">"
}

func mapParams(s *Scope, keys []string, vals []Value, vararg bool) (map[string]Value, error) {
	params := map[string]Value{}
	definedParams := len(keys)
	if vararg {
		definedParams--
	}
	for i := 0; i < definedParams && i < len(vals); i++ {
		val := vals[i]
		if mem, ok := val.(Member); ok {
			memval, err := mem.get()
			if err != nil {
				return params, err
			}
			val = memval
		}
		params[keys[i]] = val
	}
	if vararg && len(vals) > len(params)-1 {
		tbl, _ := TableClass.New(s, vals[len(keys)-1:]...)
		params[keys[len(keys)-1]] = tbl
	}
	return params, nil
}
