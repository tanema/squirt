package runtime

type Instance struct {
	class *Class
	data  map[Value]Value
}

func create(s *Scope, name string, args ...Value) (*Instance, error) {
	class, err := findClass(s, name)
	if err != nil {
		return nil, err
	}
	return class.New(s, args...)
}

func (inst *Instance) OpIndex(scope *Scope, key Value) (Value, error) {
	attr, _ := inst.class.index(scope, key, inst, false)
	if attr == nil {
		if indexattr, _ := inst.class.index(scope, "__index", inst, true); indexattr != nil {
			return indexattr.call(scope, inst, []Value{key})
		}
	}
	return inst.class.get(scope, key, inst, false)
}

func (inst *Instance) OpAssignIndex(scope *Scope, key, val Value) (Value, error) {
	attr, _ := inst.class.index(scope, key, inst, false)
	if attr == nil {
		if indexattr, _ := inst.class.index(scope, "__assignindex", inst, true); indexattr != nil {
			return indexattr.call(scope, inst, []Value{key, val})
		}
	}
	return inst.class.set(scope, key, val, inst, false)
}

func (i *Instance) Self() CVal {
	return &instanceSelf{i}
}

func (i *Instance) Super(s *Scope, self CVal, key Value, args []Value) Value {
	return i.class.Super(s, i, key, args)
}

func (i *Instance) Type() string {
	return i.class.name
}

func (i *Instance) IsA(other string) bool {
	class := i.class
	for {
		if class.name == other {
			return true
		}
		if class.parent == nil {
			return false
		}
		class = class.parent
	}
}

func (i *Instance) Op(method string, s *Scope, args ...Value) (Value, error) {
	attr, err := i.class.index(s, method, i, true)
	if err != nil {
		return nil, err
	}
	return attr.call(s, i, args)
}

func (i *Instance) ToBoolean(s *Scope) bool {
	if val, err := i.Op("tobool", s); err == nil {
		if inst := val.(*Instance); inst.IsA("Boolean") {
			return inst.data["_val"].(bool)
		}
	}
	return true
}

func (i *Instance) ToString(s *Scope) string {
	if val, err := i.Op("tostring", s); err == nil {
		if inst := val.(*Instance); inst.IsA("String") {
			return inst.data["_val"].(string)
		}
	}
	return "#<Instance of " + i.class.name + ">"
}

func (i *Instance) Error() string {
	if i.IsA("Error") {
		str, _ := i.data["message"].(string)
		return str
	}
	return ""
}
