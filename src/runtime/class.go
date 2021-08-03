package runtime

import "fmt"

type Class struct {
	name       string
	parent     *Class
	attributes map[Value]*Attribute
}

func CreateClass(name string, parent *Class, attrs ...*Attribute) *Class {
	class := &Class{
		name:       name,
		parent:     parent,
		attributes: map[Value]*Attribute{},
	}
	for _, attr := range attrs {
		if fn, ok := attr.val.(*Func); ok {
			fn.ClassName = name
		}
		class.attributes[attr.name] = attr
	}
	return class
}

func findClass(scope *Scope, className string) (*Class, error) {
	cls := scope.Get(className)
	if cls == nil {
		return nil, fmt.Errorf("undefined parent class %v", className)
	}
	final, isClass := cls.(*Class)
	if !isClass {
		return nil, fmt.Errorf("cannot use a %v as a parent class", typeOf(final))
	}
	return final, nil
}

func (class *Class) New(scope *Scope, args ...Value) (*Instance, error) {
	newInst := &Instance{class: class, data: map[Value]Value{}}
	if constructor, _ := class.index(scope, "new", newInst, true); constructor != nil {
		if _, err := constructor.call(scope, newInst, args); err != nil {
			return nil, err
		}
	} else if len(args) > 0 {
		for _, opts := range args {
			if inst, is := opts.(*Instance); is && inst.IsA("Table") {
				tbl := inst.data["_tbl"].(*Table)
				for i, key := range tbl.Keys {
					if _, err := class.set(scope, key, tbl.Values[i], newInst, true); err != nil {
						return nil, err
					}
				}
			}
		}
	}
	for _, attr := range class.attributes {
		if !attr.static && attr.refine.required {
			if val, err := class.get(scope, attr.name, newInst, true); err != nil {
				return nil, err
			} else if !toBool(scope, val) {
				return nil, fmt.Errorf("required attribute %v was not given a value", attr.name)
			}
		}
	}
	return newInst, nil
}

func (class *Class) index(scope *Scope, key Value, inst *Instance, allowPrivate bool) (*Attribute, error) {
	reqStatic := inst == nil
	strkey := toString(scope, key)
	if attr, ok := class.attributes[strkey]; ok {
		if !allowPrivate && attr.private {
			return nil, fmt.Errorf("tried to access private attribute %v", strkey)
		} else if attr.static == reqStatic && (allowPrivate || !attr.private) {
			return attr, nil
		}
	} else if class.parent != nil {
		return class.parent.index(scope, key, inst, allowPrivate)
	}
	return nil, fmt.Errorf("undefined attribute %v on class %v", toString(scope, key), class.name)
}

func (class *Class) ToString(s *Scope) string {
	return "#<Class " + class.name + ">"
}
func (class *Class) ToBoolean(s *Scope) bool {
	return true
}

func (class *Class) Self() CVal {
	return &classSelf{class}
}

func (class *Class) Super(s *Scope, self CVal, key Value, args []Value) Value {
	var inst *Instance
	if me, ok := self.(*Instance); ok && me != nil {
		inst = me
	}

	return &Func{
		ClassName: class.name,
		Name:      "super",
		Std:       true,
		LineNo:    -1,
		Fn: func(no *Scope, nope CVal, a []Value) (Value, error) {
			if class.parent != nil {
				super, err := class.parent.index(s, key, inst, true)
				if err == nil && super != nil {
					if len(a) == 0 {
						return super.call(s, inst, args)
					}
					return super.call(s, inst, a)
				}
				return nil, err
			}
			return nil, nil
		},
	}
}

func (class *Class) Type() string {
	return "Class"
}

func (class *Class) IsA(other string) bool {
	return other == "Class"
}

func (class *Class) Op(method string, s *Scope, args ...Value) (Value, error) {
	attr, err := class.index(s, method, nil, true)
	if err != nil {
		return nil, err
	}
	return attr.call(s, nil, args)
}

func (class *Class) get(scope *Scope, key Value, inst *Instance, allowPrivate bool) (Value, error) {
	attr, err := class.index(scope, key, inst, allowPrivate)
	if err != nil {
		return nil, err
	}
	return attr.get(scope, key, inst, allowPrivate)
}

func (class *Class) set(scope *Scope, key, val Value, inst *Instance, allowPrivate bool) (Value, error) {
	attr, err := class.index(scope, key, inst, allowPrivate)
	if err != nil {
		return nil, err
	}
	return attr.set(scope, key, val, inst, allowPrivate)
}

func (class *Class) OpIndex(scope *Scope, key Value) (Value, error) {
	return class.get(scope, key, nil, false)
}
func (class *Class) OpAssignIndex(scope *Scope, key, val Value) (Value, error) {
	return class.set(scope, key, val, nil, false)
}
