package runtime

import "fmt"

var NumberClass = CreateClass("Number", nil,
	Attr("_val", float64(0), nil),
	FnAttr("new", func(s *Scope, self CVal, args []Value) (Value, error) {
		if len(args) > 0 {
			self.(*Instance).data["_val"] = toNumber(args[0])
		}
		return nil, nil
	}),
	FnAttr("__add", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			return me + you, nil
		}
		return nil, fmt.Errorf("cannot add number and %v", other.Type())
	}),
	FnAttr("__sub", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			return me - you, nil
		}
		return nil, fmt.Errorf("cannot subtract number and %v", other.Type())
	}),
	FnAttr("__shiftright", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			return int(me) >> int(you), nil
		}
		return nil, fmt.Errorf("cannot shift number and-%v", other.Type())
	}),
	FnAttr("__shiftleft", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			return int(me) << int(you), nil
		}
		return nil, fmt.Errorf("cannot shift number and %v", other.Type())
	}),
	FnAttr("__and", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			return int(me) & int(you), nil
		}
		return nil, fmt.Errorf("cannot and number and %v", other.Type())
	}),
	FnAttr("__xor", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			return int(me) ^ int(you), nil
		}
		return nil, fmt.Errorf("cannot xor number and %v", other.Type())
	}),
	FnAttr("__or", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			return int(me) | int(you), nil
		}
		return nil, fmt.Errorf("cannot or number and %v", other.Type())
	}),
	FnAttr("__mul", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			return me * you, nil
		}
		return nil, fmt.Errorf("cannot mul number and %v", other.Type())
	}),
	FnAttr("__div", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			return me / you, nil // TODO divide by 0
		}
		return nil, fmt.Errorf("cannot div number and %v", other.Type())
	}),
	FnAttr("__mod", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			return int(me) % int(you), nil
		}
		return nil, fmt.Errorf("cannot mod number and %v", other.Type())
	}),
	FnAttr("__exp", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			return int(me) ^ int(you), nil
		}
		return nil, fmt.Errorf("cannot exp number and %v", other.Type())
	}),
	FnAttr("__compare", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			if me < you {
				return -1, nil
			} else if me == you {
				return 0, nil
			}
			return 1, nil
		}
		return nil, fmt.Errorf("cannot compare number and %v", other.Type())
	}),
	FnAttr("__eq", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Number") {
			me := self.(*Instance).data["_val"].(float64)
			you := other.(*Instance).data["_val"].(float64)
			return me == you, nil
		}
		return nil, fmt.Errorf("cannot mod number and %v", other.Type())
	}),
	FnAttr("tobool", func(s *Scope, self CVal, args []Value) (Value, error) {
		me := self.(*Instance).data["_val"].(float64)
		return me != 0, nil
	}),
	FnAttr("tostring", func(s *Scope, self CVal, args []Value) (Value, error) {
		me := self.(*Instance).data["_val"].(float64)
		return fmt.Sprintf("%g", me), nil
	}),
)
