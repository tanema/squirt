package runtime

type instanceSelf struct{ *Instance }

func (self *instanceSelf) OpIndex(scope *Scope, key Value) (Value, error) {
	attr, _ := self.class.index(scope, key, self.Instance, true)
	if attr == nil {
		if indexattr, _ := self.class.index(scope, "__index", self.Instance, true); indexattr != nil {
			return indexattr.call(scope, self.Instance, []Value{key})
		}
	}
	return self.class.get(scope, key, self.Instance, true)
}

func (self *instanceSelf) OpAssignIndex(scope *Scope, key, val Value) (Value, error) {
	attr, _ := self.class.index(scope, key, self.Instance, true)
	if attr == nil {
		if indexattr, _ := self.class.index(scope, "__assignindex", self.Instance, true); indexattr != nil {
			return indexattr.call(scope, self.Instance, []Value{key, val})
		}
	}
	return self.class.set(scope, key, val, self.Instance, true)
}
