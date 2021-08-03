package runtime

var BooleanClass = CreateClass("Boolean", nil,
	Attr("_val", false, nil),
	FnAttr("new", func(s *Scope, self CVal, args []Value) (Value, error) {
		if len(args) > 0 {
			self.(*Instance).data["_val"] = toBool(s, args[0])
		}
		return nil, nil
	}),
	FnAttr("__eq", func(s *Scope, self CVal, args []Value) (Value, error) {
		return self.(*Instance).data["_val"].(bool) == toBool(s, args[0]), nil
	}),
	FnAttr("tobool", func(s *Scope, self CVal, args []Value) (Value, error) {
		return self, nil
	}),
	FnAttr("tostring", func(s *Scope, self CVal, args []Value) (Value, error) {
		bl := self.(*Instance).data["_val"].(bool)
		if bl {
			return "true", nil
		}
		return "false", nil
	}),
)
