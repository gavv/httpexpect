package httpexpect

// Value provides methods to inspect attached interface{} object
// (Go representation of arbitrary JSON value) and cast it to
// concrete type.
type Value struct {
	checker Checker
	value   interface{}
}

// NewValue returns a new Value given a checker used to report failures
// and value to be inspected.
//
// checker should not be nil, but value may be nil.
//
// Example:
//  value := NewValue(checker, map[string]interface{}{"foo": 123})
//  value.Object()
//
//  value := NewValue(checker, []interface{}{"foo", 123})
//  value.Array()
//
//  value := NewValue(checker, "foo")
//  value.String()
//
//  value := NewValue(checker, 123)
//  value.Number()
//
//  value := NewValue(checker, true)
//  value.Boolean()
//
//  value := NewValue(checker, nil)
//  value.Null()
func NewValue(checker Checker, value interface{}) *Value {
	return &Value{checker, value}
}

// Raw returns underlying value attached to Value.
// This is the value originally passed to NewValue.
//
// Example:
//  value := NewValue(checker, "foo")
//  assert.Equal(t, "foo", number.Raw().(string))
func (v *Value) Raw() interface{} {
	return v.value
}

// Object returns a new Object attached to underlying value.
//
// If underlying value is not an object (map[string]interface{}), failure is reported
// and empty (but non-nil) value is returned.
//
// Example:
//  value := NewValue(checker, map[string]interface{}{"foo": 123})
//  value.Object().ContainsKey("foo")
func (v *Value) Object() *Object {
	data, ok := canonMap(v.checker, v.value)
	if !ok {
		v.checker.Fail("can't convert value to object")
	}
	return NewObject(v.checker.Clone(), data)
}

// Array returns a new Array attached to underlying value.
//
// If underlying value is not an array ([]interface{}), failure is reported and empty
// (but non-nil) value is returned.
//
// Example:
//  value := NewValue(checker, []interface{}{"foo", 123})
//  value.Array().Elements("foo", 123)
func (v *Value) Array() *Array {
	data, ok := canonArray(v.checker, v.value)
	if !ok {
		v.checker.Fail("can't convert value to array")
	}
	return NewArray(v.checker.Clone(), data)
}

// String returns a new String attached to underlying value.
//
// If underlying value is not string, failure is reported and empty (but non-nil)
// value is returned.
//
// Example:
//  value := NewValue(checker, "foo")
//  value.String().EqualFold("FOO")
func (v *Value) String() *String {
	data, ok := v.value.(string)
	if !ok {
		v.checker.Fail("can't convert value to string")
	}
	return NewString(v.checker.Clone(), data)
}

// Number returns a new Number attached to underlying value.
//
// If underlying value is not a number (numeric type convertible to float64), failure
// is reported and empty (but non-nil) value is returned.
//
// Example:
//  value := NewValue(checker, 123)
//  value.Number().InRange(100, 200)
func (v *Value) Number() *Number {
	data, ok := canonNumber(v.checker, v.value)
	if !ok {
		v.checker.Fail("can't convert value to number")
	}
	return NewNumber(v.checker.Clone(), data)
}

// Boolean returns a new Boolean attached to underlying value.
//
// If underlying value is not a bool, failure is reported and empty (but non-nil)
// value is returned.
//
// Example:
//  value := NewValue(checker, true)
//  value.Boolean().True()
func (v *Value) Boolean() *Boolean {
	data, ok := v.value.(bool)
	if !ok {
		v.checker.Fail("can't convert value to boolean")
	}
	return NewBoolean(v.checker.Clone(), data)
}

// Null succeedes if value is nil.
//
// Note that non-nil interface{} that points to nil value (e.g. nil slice or map)
// is also treated as null value. Empty (non-nil) slice or map, empty string, and
// zero number are not treated as null value.
//
// Example:
//  value := NewValue(checker, nil)
//  value.Null()
//
//  value := NewValue(checker, []interface{}(nil))
//  value.Null()
func (v *Value) Null() *Value {
	data, ok := canonValue(v.checker, v.value)
	if !ok {
		return v
	}
	if data != nil {
		v.checker.Fail("expected nil value, got %v", v.value)
	}
	return v
}

// NotNull succeedes if value is not nil.
//
// Note that non-nil interface{} that points to nil value (e.g. nil slice or map)
// is also treated as null value. Empty (non-nil) slice or map, empty string, and
// zero number are not treated as null value.
//
// Example:
//  value := NewValue(checker, "")
//  value.NotNull()
//
//  value := NewValue(checker, make([]interface{}, 0)
//  value.Null()
func (v *Value) NotNull() *Value {
	data, ok := canonValue(v.checker, v.value)
	if !ok {
		return v
	}
	if data == nil {
		v.checker.Fail("expected non-nil value, got %v", v.value)
	}
	return v
}
