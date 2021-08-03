package runtime

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/tanema/squirt/src/lang"
)

type (
	Attribute struct {
		name    string
		val     Value
		private bool
		static  bool
		refine  *Refinement
	}
	Refinement struct {
		constant bool
		class    string
		required bool
		get      string
		set      string
	}
)

func Fn(name string, fn FnSig) *Func {
	return &Func{Name: name, Fn: fn, Std: true, LineNo: -1}
}

func FnAttr(name string, fn FnSig) *Attribute {
	return Attr(name, Fn(name, fn), &Refinement{constant: true})
}

func Attr(name string, val Value, refinements *Refinement) *Attribute {
	private := strings.HasPrefix(name, "_")
	static := (private && unicode.IsUpper(rune(name[1]))) || unicode.IsUpper(rune(name[0]))
	if refinements == nil {
		refinements = &Refinement{}
	}
	return &Attribute{
		name:    name,
		val:     val,
		private: private,
		static:  static,
		refine:  refinements,
	}
}

func (attr *Attribute) call(s *Scope, self *Instance, args []Value) (Value, error) {
	if fn, is := attr.val.(*Func); is {
		return fn.call(s, self, args)
	}
	return nil, fmt.Errorf("tried to call a non callable object (%v)", typeOf(attr.val))
}

func (attr *Attribute) get(scope *Scope, key Value, inst *Instance, allowPrivate bool) (Value, error) {
	if inst != nil {
		if !allowPrivate && attr.refine.get != "" {
			return inst.Op(attr.refine.get, scope)
		}
		if val, ok := inst.data[toString(scope, key)]; ok {
			return val, nil
		}
	}
	return attr.val, nil
}

func (attr *Attribute) set(scope *Scope, key, val Value, inst *Instance, allowPrivate bool) (Value, error) {
	if attr.refine.constant {
		return nil, fmt.Errorf("cannot assign to constant attribute %v", attr.name)
	}
	if v, ok := val.(CVal); ok && attr.refine.class != "" {
		if !v.IsA(attr.refine.class) {
			return nil, fmt.Errorf("incorrect type %v passed to attribute %v", v.Type(), attr.name)
		}
	}
	if inst != nil {
		if !allowPrivate && attr.refine.set != "" {
			return inst.Op(attr.refine.set, scope, val)
		}
		strkey := toString(scope, key)
		inst.data[strkey] = val
		return inst.data[strkey], nil
	}
	attr.val = val
	return attr.val, nil
}

func parseRefinement(scope *Scope, runtime *Runtime, cond *lang.Object) (*Refinement, error) {
	if cond == nil {
		return nil, nil
	}
	tbl, err := runtime.evalTable(scope, *cond)
	if err != nil {
		return nil, err
	}

	refine := &Refinement{}
	for i, key := range tbl.Keys {
		name, isStrName := key.(*Instance).data["_val"].(string)
		if !isStrName {
			continue
		}
		switch name {
		case "const":
			refine.constant = toBool(scope, tbl.Values[i])
		case "type":
			if inst, is := tbl.Values[i].(*Instance); is && inst.IsA("String") {
				refine.class = inst.data["_val"].(string)
			} else if cls, is := tbl.Values[i].(*Class); is {
				refine.class = cls.name
			} else {
				return nil, fmt.Errorf("invalid value provided to type refinement")
			}
		case "required":
			refine.required = toBool(scope, tbl.Values[i])
		case "get":
			if inst, is := tbl.Values[i].(*Instance); is && inst.IsA("String") {
				refine.get = inst.data["_val"].(string)
			} else if fn, is := tbl.Values[i].(*Func); is {
				refine.get = fn.Name
			} else {
				return nil, fmt.Errorf("invalid value provided to get refinement")
			}
		case "set":
			if inst, is := tbl.Values[i].(*Instance); is && inst.IsA("String") {
				refine.set = inst.data["_val"].(string)
			} else if fn, is := tbl.Values[i].(*Func); is {
				refine.set = fn.Name
			} else {
				return nil, fmt.Errorf("invalid value provided to set refinement")
			}
		default:
			return nil, fmt.Errorf("invalid refinement %v", toString(scope, key))
		}
	}
	return refine, nil
}
