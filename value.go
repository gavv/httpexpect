package httpexpect

// Value provides methods to inspect attached interface{} object
// (Go representation of arbitrary JSON value) and cast it to
// concrete type.
type Value struct {
	chain chain
	value interface{}
}

// NewValue returns a new Value given a reporter used to report failures
// and value to be inspected.
//
// reporter should not be nil, but value may be nil.
//
// Example:
//  value := NewValue(t, map[string]interface{}{"foo": 123})
//  value.Object()
//
//  value := NewValue(t, []interface{}{"foo", 123})
//  value.Array()
//
//  value := NewValue(t, "foo")
//  value.String()
//
//  value := NewValue(t, 123)
//  value.Number()
//
//  value := NewValue(t, true)
//  value.Boolean()
//
//  value := NewValue(t, nil)
//  value.Null()
func NewValue(reporter Reporter, value interface{}) *Value {
	return &Value{makeChain(reporter), value}
}

// Raw returns underlying value attached to Value.
// This is the value originally passed to NewValue.
//
// Example:
//  value := NewValue(t, "foo")
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
//  value := NewValue(t, map[string]interface{}{"foo": 123})
//  value.Object().ContainsKey("foo")
func (v *Value) Object() *Object {
	data, ok := canonMap(&v.chain, v.value)
	if !ok {
		v.chain.fail("\nexpected object value (map or struct), but got:\n%s",
			dumpValue(v.value))
	}
	return &Object{v.chain, data}
}

// Array returns a new Array attached to underlying value.
//
// If underlying value is not an array ([]interface{}), failure is reported and empty
// (but non-nil) value is returned.
//
// Example:
//  value := NewValue(t, []interface{}{"foo", 123})
//  value.Array().Elements("foo", 123)
func (v *Value) Array() *Array {
	data, ok := canonArray(&v.chain, v.value)
	if !ok {
		v.chain.fail("\nexpected array value, but got:\n%s",
			dumpValue(v.value))
	}
	return &Array{v.chain, data}
}

// String returns a new String attached to underlying value.
//
// If underlying value is not string, failure is reported and empty (but non-nil)
// value is returned.
//
// Example:
//  value := NewValue(t, "foo")
//  value.String().EqualFold("FOO")
func (v *Value) String() *String {
	data, ok := v.value.(string)
	if !ok {
		v.chain.fail("\nexpected string value, but got:\n%s",
			dumpValue(v.value))
	}
	return &String{v.chain, data}
}

// Number returns a new Number attached to underlying value.
//
// If underlying value is not a number (numeric type convertible to float64), failure
// is reported and empty (but non-nil) value is returned.
//
// Example:
//  value := NewValue(t, 123)
//  value.Number().InRange(100, 200)
func (v *Value) Number() *Number {
	data, ok := canonNumber(&v.chain, v.value)
	if !ok {
		v.chain.fail("\nexpected numeric value, but got:\n%s",
			dumpValue(v.value))
	}
	return &Number{v.chain, data}
}

// Boolean returns a new Boolean attached to underlying value.
//
// If underlying value is not a bool, failure is reported and empty (but non-nil)
// value is returned.
//
// Example:
//  value := NewValue(t, true)
//  value.Boolean().True()
func (v *Value) Boolean() *Boolean {
	data, ok := v.value.(bool)
	if !ok {
		v.chain.fail("\nexpected boolean value, but got:\n%s",
			dumpValue(v.value))
	}
	return &Boolean{v.chain, data}
}

// Null succeedes if value is nil.
//
// Note that non-nil interface{} that points to nil value (e.g. nil slice or map)
// is also treated as null value. Empty (non-nil) slice or map, empty string, and
// zero number are not treated as null value.
//
// Example:
//  value := NewValue(t, nil)
//  value.Null()
//
//  value := NewValue(t, []interface{}(nil))
//  value.Null()
func (v *Value) Null() *Value {
	data, ok := canonValue(&v.chain, v.value)
	if !ok {
		return v
	}
	if data != nil {
		v.chain.fail("\nexpected nil value, but got:\n%s",
			dumpValue(v.value))
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
//  value := NewValue(t, "")
//  value.NotNull()
//
//  value := NewValue(t, make([]interface{}, 0)
//  value.Null()
func (v *Value) NotNull() *Value {
	data, ok := canonValue(&v.chain, v.value)
	if !ok {
		return v
	}
	if data == nil {
		v.chain.fail("\nexpected non-nil value, but got:\n%s",
			dumpValue(v.value))
	}
	return v
}
