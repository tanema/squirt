# Declaring class
```
class Dog isa Animal do
  attr Leg_count = 4 // Class variable (Dog.Leg_count)
  attr tail_count = 4, {required: true} // normal variable, not instance

  func new(args) // constructor
    super() // super auto sets properties
  end

  // uppercase indicates Class method
  func Bark()
    print("woof")
  end

  func sayName()
    print("hello my name is ${self.name}")
  end

  // underscore prefix is private?
  func _privateFunc()
  end

  // combine for private class method
  func _PrivateClassFunc()
  end

  // operator override special methods (two underscores so its private and instance)
  func __add(other)
    return new(Dog, {name: "${self.name} + ${other.name}"})
  end
end

Dog.Bark()
dog = new(Dog, {name: "spot"})
dog.sayName()
dog.name = "Spotty"
dog.sayName()
print(dog.name)
```

# property refinements
- required
- type
- get/set
- const

# Operator Overloading
- `__len`
- `__iter`
- `__string`
- `__add`
- `__sub`
- `__div`
- `__mul`
- `__exp`
- `__mod`
- `__shiftleft`
- `__shiftright`
- `__and`
- `__xor`
- `__or`
- `__lt`
- `__lte`
- `__gt`
- `__gte`
- `__eq`
- `__delete`
- `__index`
- `__index=`
- `tostring`
- `tobool`
