package runtime

import "io"

// Scope captures all definitions and their values
type Scope struct {
	data  map[string]Value
	out   io.StringWriter
	outer *Scope
}

func newScope(outer *Scope, binds map[string]Value, out io.StringWriter) *Scope {
	if binds == nil {
		binds = map[string]Value{}
	}
	return &Scope{
		data:  binds,
		out:   out,
		outer: outer,
	}
}

// Child creates a new Scope that inherits this one.
func (scope *Scope) Child(binds map[string]Value) *Scope {
	return newScope(scope, binds, scope.out)
}

func (scope *Scope) find(key string) *Scope {
	if _, ok := scope.data[key]; ok {
		return scope
	} else if scope.outer == nil {
		return nil
	}
	return scope.outer.find(key)
}

// Set will set the definition of a name on the current Scope
func (scope *Scope) Set(key string, value Value) {
	if found := scope.find(key); found != nil {
		found.data[key] = value
	} else {
		scope.data[key] = value
	}
}

// Get will retreive the value of a name recursively up the parentage of this scope
func (scope *Scope) Get(key string) Value {
	if found := scope.find(key); found != nil {
		return found.data[key]
	}
	return nil
}
