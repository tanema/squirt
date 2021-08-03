package runtime

import "fmt"

var ErrorClass = CreateClass("Error", nil,
	Attr("message", "an error has occurred", nil),
	FnAttr("new", func(s *Scope, self CVal, args []Value) (Value, error) {
		if len(args) > 0 {
			self.(*Instance).data["message"] = toString(s, args[0])
		}
		return nil, nil
	}),
	FnAttr("__eq", func(s *Scope, self CVal, args []Value) (Value, error) {
		other := args[0].(CVal)
		msg, _ := self.(*Instance).data["message"]
		if other.IsA("Error") {
			othermsg, _ := other.(*Instance).data["message"]
			return msg.(string) == othermsg.(string), nil
		}
		return false, nil
	}),
	FnAttr("tobool", func(s *Scope, self CVal, args []Value) (Value, error) {
		return true, nil
	}),
	FnAttr("tostring", func(s *Scope, self CVal, args []Value) (Value, error) {
		inst := self.(*Instance)
		return fmt.Sprintf("%v: %v", inst.class.name, inst.data["message"].(string)), nil
	}),
)

func createErr(s *Scope, cls *Class, msg string) (Value, error) {
	err, _ := cls.New(s, msg)
	return nil, err
}
