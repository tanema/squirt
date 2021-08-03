package runtime

import (
	"fmt"
)

var (
	registeredLibs = map[string]func(*Scope) (Value, error){}
	requireCache   = map[string]Value{}
)

func stdRequire(s *Scope, self CVal, a []Value) (Value, error) {
	if len(a) == 0 {
		return createErr(s, ArgumentError, "not enough arguments to require")
	} else if inst, ok := a[0].(*Instance); !ok || !inst.IsA("String") {
		return nil, fmt.Errorf("wrong value type passed to require")
	} else {
		return RequirePath(s, inst.data["_val"].(string))
	}
}

func RegisterLib(name string, val func(*Scope) (Value, error)) {
	registeredLibs[name] = val
}

func RequirePath(s *Scope, path string) (Value, error) {
	if val, ok := requireCache[path]; ok {
		println("hit")
		return val, nil
	} else if fn, ok := registeredLibs[path]; ok {
		val, err := fn(s)
		if err != nil {
			return nil, err
		}
		requireCache[path] = val
		return val, nil
	}

	res, err := EvalFile(s.Child(map[string]Value{}), path)
	if ret, ok := res.(Return); ok {
		return ret.Vals, err
	}
	return nil, err
}
