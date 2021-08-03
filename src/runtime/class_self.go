package runtime

type classSelf struct{ *Class }

func (class *classSelf) OpIndex(scope *Scope, key Value) (Value, error) {
	return class.get(scope, key, nil, true)
}
func (class *classSelf) OpAssignIndex(scope *Scope, key, val Value) (Value, error) {
	return class.set(scope, key, val, nil, false)
}
