package runtime

import "fmt"

var StringClass = CreateClass("String", nil,
	Attr("_val", "", nil),
	FnAttr("new", func(s *Scope, self CVal, args []Value) (Value, error) {
		if len(args) > 0 {
			self.(*Instance).data["_val"] = toString(s, args[0])
		}
		return nil, nil
	}),
	FnAttr("__index", func(s *Scope, self CVal, args []Value) (Value, error) {
		val := self.(*Instance).data["_val"].(string)
		if rng, isRange := args[0].(Range); isRange && (rng.Start > len(val) || rng.End > len(val)) {
			return nil, fmt.Errorf("range index out of range")
		} else if isRange {
			return string(val[rng.Start:rng.End]), nil
		} else if inx, isInt := isIntKey(args[0]); !isInt {
			return nil, fmt.Errorf("non int key used to index a string")
		} else if inx > len(val) {
			return nil, fmt.Errorf("index out of range")
		} else {
			return string(val[inx]), nil
		}
	}),
	FnAttr("__assignindex", func(s *Scope, self CVal, args []Value) (Value, error) {
		val := self.(*Instance).data["_val"].(string)
		var start, end string
		if rng, isRange := args[0].(Range); isRange && (rng.Start > len(val) || rng.End > len(val)) {
			return nil, fmt.Errorf("range index out of range")
		} else if isRange {
			start = val[:rng.Start]
			end = val[rng.End:]
		} else if inx, isInt := isIntKey(args[0]); !isInt {
			return nil, fmt.Errorf("non int key used to index a string")
		} else if inx > len(val) {
			return nil, fmt.Errorf("index out of range")
		} else {
			start = val[:inx]
			end = val[inx+1:]
		}
		self.(*Instance).data["_val"] = start + toString(s, args[1]) + end
		return self, nil
	}),
	FnAttr("__add", func(s *Scope, self CVal, args []Value) (Value, error) {
		return self.(*Instance).data["_val"].(string) + toString(s, args[0]), nil
	}),
	FnAttr("__eq", func(s *Scope, self CVal, args []Value) (Value, error) {
		return self.(*Instance).data["_val"].(string) == toString(s, args[0]), nil
	}),
	FnAttr("__len", func(s *Scope, self CVal, args []Value) (Value, error) {
		return len(self.(*Instance).data["_val"].(string)), nil
	}),
	FnAttr("tobool", func(s *Scope, self CVal, args []Value) (Value, error) {
		return self.(*Instance).data["_val"].(string) != "", nil
	}),
	FnAttr("tostring", func(s *Scope, self CVal, args []Value) (Value, error) {
		return self, nil
	}),
)
