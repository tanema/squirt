package runtime

import "testing"

func TestOperators(t *testing.T) {
	var inst interface{} = &Instance{}
	_ = inst.(CVal)
	var iself interface{} = &instanceSelf{}
	_ = iself.(CVal)
	var cls interface{} = &Class{}
	_ = cls.(CVal)
	var cself interface{} = &classSelf{}
	_ = cself.(CVal)
	var fn interface{} = &Func{}
	_ = fn.(CVal)
}
