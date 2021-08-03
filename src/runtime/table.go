package runtime

import (
	"fmt"
	"strings"
)

type (
	Table struct {
		Arr    []Value
		Keys   []Value
		Values []Value
	}
)

var TableClass = CreateClass("Table", nil,
	Attr("_tbl", nil, nil),
	FnAttr("new", func(s *Scope, self CVal, args []Value) (Value, error) {
		if len(args) == 1 {
			if tbl, ok := args[0].(*Table); ok {
				self.(*Instance).data["_tbl"] = tbl
				return nil, nil
			} else if tbl, ok := args[0].(map[string]Value); ok {
				newTbl := &Table{}
				for key, val := range tbl {
					instKey, _ := ToValue(s, key)
					newTbl.Keys = append(newTbl.Keys, instKey)
					newTbl.Values = append(newTbl.Values, val)
				}
				self.(*Instance).data["_tbl"] = newTbl
				return nil, nil
			} else if tbl, ok := args[0].(CVal); ok && tbl.IsA("Table") {
				self.(*Instance).data["_tbl"] = tbl.(*Instance).data["_tbl"].(*Table)
				return nil, nil
			}
		}
		self.(*Instance).data["_tbl"] = &Table{Arr: args}
		return nil, nil
	}),
	FnAttr("__index", func(s *Scope, self CVal, args []Value) (Value, error) {
		tbl := self.(*Instance).data["_tbl"].(*Table)
		if rng, isRange := args[0].(Range); isRange {
			return &Table{Arr: tbl.Arr[min(rng.Start, len(tbl.Arr)):min(rng.End, len(tbl.Arr))]}, nil
		} else if i, itis := isIntKey(args[0]); itis {
			if i < len(tbl.Arr) {
				return tbl.Arr[i], nil
			}
			return nil, nil
		}
		_, val := tbl.findKey(s, args[0])
		return val, nil
	}),
	FnAttr("__assignindex", func(s *Scope, self CVal, args []Value) (Value, error) {
		tbl := self.(*Instance).data["_tbl"].(*Table)
		if rng, isRange := args[0].(Range); isRange {
			tbl.ensureSize(s, rng.End-1)
			start, end := tbl.Arr[:rng.Start], tbl.Arr[rng.End:]
			if inst, is := args[1].(*Instance); is && inst.IsA("Table") {
				other := inst.data["_tbl"].(*Table)
				tbl.Arr = append(start, other.Arr...)
			} else {
				tbl.Arr = append(start, args[1])
			}
			if len(end) > 0 {
				tbl.Arr = append(tbl.Arr, end...)
			}
			return args[1], nil
		}
		return tbl.assign(s, args[0], args[1])
	}),
	FnAttr("__add", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Table") {
			newTable := &Table{}
			newTable.add(s, self.(*Instance).data["_tbl"].(*Table))
			newTable.add(s, other.(*Instance).data["_tbl"].(*Table))
			return newTable, nil
		}
		return nil, fmt.Errorf("cannot add table and %v", other.Type())
	}),
	FnAttr("__sub", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		if other.IsA("Table") {
			newTable := &Table{}
			selftbl := self.(*Instance).data["_tbl"].(*Table)
			othertbl := other.(*Instance).data["_tbl"].(*Table)
			for _, v := range selftbl.Arr {
				if i := othertbl.findValue(s, v); i == -1 {
					newTable.Arr = append(newTable.Arr, v)
				}
			}
			for i, key := range selftbl.Keys {
				if _, val := othertbl.findKey(s, key); val == nil {
					newTable.assign(s, key, selftbl.Values[i])
				}
			}
			return newTable, nil
		}
		return nil, fmt.Errorf("cannot sub table and %v", other.Type())
	}),
	FnAttr("__shiftleft", func(s *Scope, self CVal, args []Value) (Value, error) {
		tbl := self.(*Instance).data["_tbl"].(*Table)
		tbl.Arr = append(tbl.Arr, args[0])
		return self, nil
	}),
	FnAttr("__eq", func(s *Scope, self CVal, args []Value) (Value, error) {
		tbl := self.(*Instance).data["_tbl"].(*Table)
		other := args[0].(CVal)
		if other.IsA("Table") {
			othertbl := other.(*Instance).data["_tbl"].(*Table)
			return tbl == othertbl, nil
		}
		return false, nil
	}),
	FnAttr("__len", func(s *Scope, self CVal, args []Value) (Value, error) {
		return len(self.(*Instance).data["_tbl"].(*Table).Arr), nil
	}),
	FnAttr("__del", func(s *Scope, self CVal, args []Value) (Value, error) {
		tbl := self.(*Instance).data["_tbl"].(*Table)
		if i, is := isIntKey(args[0]); is {
			if int(i) >= len(tbl.Arr) {
				return nil, nil
			}
			val := tbl.Arr[i]
			tbl.Arr = append(tbl.Arr[:i], tbl.Arr[i+1:]...)
			return val, nil
		}

		if i, val := tbl.findKey(s, args[0]); val != nil {
			tbl.Keys = append(tbl.Keys[:i], tbl.Keys[i+1:]...)
			tbl.Values = append(tbl.Values[:i], tbl.Values[i+1:]...)
			return val, nil
		}

		return nil, nil
	}),
	FnAttr("tobool", func(s *Scope, self CVal, args []Value) (Value, error) {
		return true, nil
	}),
	FnAttr("tostring", func(s *Scope, self CVal, args []Value) (Value, error) {
		tbl := self.(*Instance).data["_tbl"].(*Table)
		strList := []string{}
		for _, e := range tbl.Arr {
			strList = append(strList, toString(s, e))
		}
		for i, key := range tbl.Keys {
			strList = append(strList, toString(s, key)+": "+toString(s, tbl.Values[i]))
		}
		return "{" + strings.Join(strList, ", ") + "}", nil
	}),
)

func (tbl *Table) add(s *Scope, other *Table) {
	tbl.Arr = append(tbl.Arr, other.Arr...)
	for i, key := range other.Keys {
		tbl.assign(s, key, other.Values[i])
	}
}

func (tbl *Table) ensureSize(s *Scope, size int) {
	if size < len(tbl.Arr) {
		return
	}
	ext := size - len(tbl.Arr)
	for i := 0; i <= ext; i++ {
		n, _ := ToValue(s, nil)
		tbl.Arr = append(tbl.Arr, n)
	}
}

func (tbl *Table) assign(s *Scope, key, val Value) (Value, error) {
	if i, is := isIntKey(key); is {
		tbl.ensureSize(s, i)
		tbl.Arr[i] = val
		return val, nil
	}

	if i, v := tbl.findKey(s, key); v == nil {
		tbl.Keys = append(tbl.Keys, key)
		tbl.Values = append(tbl.Values, val)
	} else {
		tbl.Values[i] = val
	}
	return val, nil
}

func (tbl *Table) findKey(s *Scope, key Value) (int, Value) {
	for i, k := range tbl.Keys {
		if key == k {
			return i, tbl.Values[i]
		}
		if inst, ok := k.(CVal); ok {
			if res, err := inst.Op("__eq", s, key); err == nil {
				if toBool(s, res) {
					return i, tbl.Values[i]
				}
			}
		}
	}
	return -1, nil
}

func (tbl *Table) findValue(s *Scope, val Value) int {
	for i, v := range tbl.Arr {
		if val == v {
			return i
		}
		if inst, ok := val.(CVal); ok {
			if res, err := inst.Op("__eq", s, val); err == nil {
				if toBool(s, res) {
					return i
				}
			}
		}
	}
	return -1
}

func isIntKey(key Value) (int, bool) {
	if inst, ok := key.(CVal); ok {
		if inst.IsA("Number") {
			val := inst.(*Instance).data["_val"].(float64)
			if float64(int(val)) == val && val >= 0 {
				return int(val), true
			}
		}
	}
	return -1, false
}

func max(x, y int) int {
	if x > y {
		return x
	}
	return y
}

func min(x, y int) int {
	if x < y {
		return x
	}
	return y
}
