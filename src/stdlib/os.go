package stdlib

import (
	rt "runtime"
	"time"

	"github.com/tanema/squirt/src/runtime"
)

func OSLib(scope *runtime.Scope) (runtime.Value, error) {
	arch, _ := runtime.ToValue(scope, rt.GOARCH)
	os, _ := runtime.ToValue(scope, rt.GOOS)

	return runtime.ToValue(scope, map[string]runtime.Value{
		"Arch": arch,
		"OS":   os,
		"time": runtime.Fn("time", osTime),
	})
}

func osTime(s *runtime.Scope, self runtime.CVal, args []runtime.Value) (runtime.Value, error) {
	return time.Now().Unix(), nil
}
