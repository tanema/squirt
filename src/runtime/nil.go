package runtime

var NilClass = CreateClass("Nil", nil,
	FnAttr("__eq", func(s *Scope, self CVal, args []Value) (Value, error) {
		return args[0].(CVal).IsA("Nil"), nil
	}),
	FnAttr("tobool", func(s *Scope, self CVal, args []Value) (Value, error) {
		return false, nil
	}),
	FnAttr("tostring", func(s *Scope, self CVal, args []Value) (Value, error) {
		return "nil", nil
	}),
)
