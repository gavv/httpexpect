package httpexpect

type Value struct {
	checker Checker
	value   interface{}
}

func NewValue(checker Checker, value interface{}) *Value {
	return &Value{checker, value}
}

func (v *Value) Raw() interface{} {
	return v.value
}

func (v *Value) Object() *Object {
	data, ok := canonMap(v.checker, v.value)
	if !ok {
		v.checker.Fail("can't convert value to object")
	}
	return NewObject(v.checker.Clone(), data)
}

func (v *Value) Array() *Array {
	data, ok := canonArray(v.checker, v.value)
	if !ok {
		v.checker.Fail("can't convert value to array")
	}
	return NewArray(v.checker.Clone(), data)
}

func (v *Value) String() *String {
	data, ok := v.value.(string)
	if !ok {
		v.checker.Fail("can't convert value to string")
	}
	return NewString(v.checker.Clone(), data)
}

func (v *Value) Number() *Number {
	data, ok := canonNumber(v.checker, v.value)
	if !ok {
		v.checker.Fail("can't convert value to number")
	}
	return NewNumber(v.checker.Clone(), data)
}

func (v *Value) Boolean() *Boolean {
	data, ok := v.value.(bool)
	if !ok {
		v.checker.Fail("can't convert value to boolean")
	}
	return NewBoolean(v.checker.Clone(), data)
}

func (v *Value) Null() {
	v.checker.Equal(v.value, nil)
}

func (v *Value) NotNull() {
	v.checker.NotEqual(v.value, nil)
}
