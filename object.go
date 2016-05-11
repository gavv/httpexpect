package httpexpect

import (
	"reflect"
)

// Object provides methods to inspect attached map[string]interface{} object
// (Go representation of JSON object).
type Object struct {
	checker Checker
	value   map[string]interface{}
}

// NewObject returns a new Object given a checker used to report failures
// and value to be inspected.
//
// Both checker and value should not be nil. If value is nil, failure is reported.
//
// Example:
//  object := NewObject(NewAssertChecker(t), map[string]interface{}{"foo": 123})
func NewObject(checker Checker, value map[string]interface{}) *Object {
	if value == nil {
		checker.Fail("expected non-nil map value")
	} else {
		value, _ = canonMap(checker, value)
	}
	return &Object{checker, value}
}

// Raw returns underlying value attached to Object.
// This is the value originally passed to NewObject, converted to canonical form.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  assert.Equal(t, map[string]interface{}{"foo": 123.0}, object.Raw())
func (o *Object) Raw() map[string]interface{} {
	return o.value
}

// Keys returns a new Array object that may be used to inspect objects keys.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123, "bar": 456})
//  object.Keys().ContainsOnly("foo", "bar")
func (o *Object) Keys() *Array {
	keys := []interface{}{}
	for k := range o.value {
		keys = append(keys, k)
	}
	return NewArray(o.checker.Clone(), keys)
}

// Values returns a new Array object that may be used to inspect objects values.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123, "bar": 456})
//  object.Values().ContainsOnly(123, 456)
func (o *Object) Values() *Array {
	values := []interface{}{}
	for _, v := range o.value {
		values = append(values, v)
	}
	return NewArray(o.checker.Clone(), values)
}

// Value returns a new Value object that may be used to inspect single value
// for given key.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.Value("foo").Number().Equal(123)
func (o *Object) Value(key string) *Value {
	value, ok := o.value[key]
	if !ok {
		o.checker.Fail("\nexpected object containing key '%s', but got:\n%s",
			key, dumpValue(o.checker, o.value))
		return NewValue(o.checker.Clone(), nil)
	}
	return NewValue(o.checker.Clone(), value)
}

// Empty succeedes if object is empty.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{})
//  object.Empty()
func (o *Object) Empty() *Object {
	return o.Equal(map[string]interface{}{})
}

// NotEmpty succeedes if object is non-empty.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.NotEmpty()
func (o *Object) NotEmpty() *Object {
	return o.NotEqual(map[string]interface{}{})
}

// Equal succeedes if object is equal to another object.
// Before comparison, both objects are converted to canonical form.
//
// value should map[string]interface{} or struct.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.Equal(map[string]interface{}{"foo": 123})
func (o *Object) Equal(value interface{}) *Object {
	expected, ok := canonMap(o.checker, value)
	if !ok {
		return o
	}
	if !reflect.DeepEqual(expected, o.value) {
		o.checker.Fail("\nexpected object equal to:\n%s\n\nbut got:\n%s\n\ndiff:\n%s",
			dumpValue(o.checker, expected),
			dumpValue(o.checker, o.value),
			diffValues(o.checker, expected, o.value))
	}
	return o
}

// NotEqual succeedes if object is not equal to another object.
// Before comparison, both objects are converted to canonical form.
//
// value should map[string]interface{} or struct.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.Equal(map[string]interface{}{"bar": 123})
func (o *Object) NotEqual(v interface{}) *Object {
	expected, ok := canonMap(o.checker, v)
	if !ok {
		return o
	}
	if reflect.DeepEqual(expected, o.value) {
		o.checker.Fail("\nexpected object NOT equal to:\n%s",
			dumpValue(o.checker, expected))
	}
	return o
}

// ContainsKey succeedes if object contains given key.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.ContainsKey("foo")
func (o *Object) ContainsKey(key string) *Object {
	if !o.containsKey(key) {
		o.checker.Fail("\nexpected object containing key '%s', but got:\n%s",
			key, dumpValue(o.checker, o.value))
	}
	return o
}

// NotContainsKey succeedes if object doesn't contain given key.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.NotContainsKey("bar")
func (o *Object) NotContainsKey(key string) *Object {
	if o.containsKey(key) {
		o.checker.Fail(
			"\nexpected object NOT containing key '%s', but got:\n%s", key,
			dumpValue(o.checker, o.value))
	}
	return o
}

// ContainsMap succeedes if object contains given sub-object.
// Before comparison, both objects are converted to canonical form.
//
// value should map[string]interface{} or struct.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{
//      "foo": 123,
//      "bar": []interface{}{"x", "y"},
//      "bar": map[string]interface{}{
//          "a": true,
//          "b": false,
//      },
//  })
//
//  object.ContainsMap(map[string]interface{}{  // success
//      "foo": 123,
//      "bar": map[string]interface{}{
//          "a": true,
//      },
//  })
//
//  object.ContainsMap(map[string]interface{}{  // failure
//      "foo": 123,
//      "qux": 456,
//  })
//
//  object.ContainsMap(map[string]interface{}{  // failure, slices should match exactly
//      "bar": []interface{}{"x"},
//  })
func (o *Object) ContainsMap(value interface{}) *Object {
	if !o.containsMap(value) {
		o.checker.Fail("\nexpected object containing sub-object:\n%s\n\nbut got:\n%s",
			dumpValue(o.checker, value), dumpValue(o.checker, o.value))
	}
	return o
}

// NotContainsMap succeedes if object doesn't contain given sub-object exactly.
// Before comparison, both objects are converted to canonical form.
//
// value should map[string]interface{} or struct.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123, "bar": 456})
//  object.NotContainsMap(map[string]interface{}{"foo": 123, "bar": "no-no-no"})
func (o *Object) NotContainsMap(value interface{}) *Object {
	if o.containsMap(value) {
		o.checker.Fail("\nexpected object NOT containing sub-object:\n%s\n\nbut got:\n%s",
			dumpValue(o.checker, value), dumpValue(o.checker, o.value))
	}
	return o
}

// ValueEqual succeedes if object's value for given key is equal to given value.
// Before comparison, both values are converted to canonical form.
//
// value should map[string]interface{} or struct.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.ValueEqual("foo", 123)
func (o *Object) ValueEqual(key string, value interface{}) *Object {
	if !o.containsKey(key) {
		o.checker.Fail("\nexpected object containing key '%s', but got:\n%s",
			key, dumpValue(o.checker, o.value))
		return o
	}
	expected, ok := canonValue(o.checker, value)
	if !ok {
		return o
	}
	if !reflect.DeepEqual(expected, o.value[key]) {
		o.checker.Fail(
			"\nexpected value for key '%s' equal to:\n%s\n\nbut got:\n%s\n\ndiff:\n%s",
			key,
			dumpValue(o.checker, expected),
			dumpValue(o.checker, o.value[key]),
			diffValues(o.checker, expected, o.value[key]))
	}
	return o
}

// ValueNotEqual succeedes if object's value for given key is not equal to given value.
// Before comparison, both values are converted to canonical form.
//
// value should map[string]interface{} or struct.
//
// If object doesn't contain any value for given key, failure is reported.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.ValueNotEqual("foo", "bad value")  // success
//  object.ValueNotEqual("bar", "bad value")  // failure! (key is missing)
func (o *Object) ValueNotEqual(key string, value interface{}) *Object {
	if !o.containsKey(key) {
		o.checker.Fail("\nexpected object containing key '%s', but got:\n%s",
			key, dumpValue(o.checker, o.value))
		return o
	}
	expected, ok := canonValue(o.checker, value)
	if !ok {
		return o
	}
	if reflect.DeepEqual(expected, o.value[key]) {
		o.checker.Fail("\nexpected value for key '%s' NOT equal to:\n%s",
			key,
			dumpValue(o.checker, expected))
	}
	return o
}

func (o *Object) containsKey(key string) bool {
	for k := range o.value {
		if k == key {
			return true
		}
	}
	return false
}

func (o *Object) containsMap(sm interface{}) bool {
	submap, ok := canonMap(o.checker, sm)
	if !ok {
		return false
	}
	return checkContainsMap(o.checker, o.value, submap)
}

func checkContainsMap(checker Checker, outer, inner map[string]interface{}) bool {
	for k, iv := range inner {
		ov, ok := outer[k]
		if !ok {
			return false
		}
		if ovm, ok := ov.(map[string]interface{}); ok {
			if ivm, ok := iv.(map[string]interface{}); ok {
				if !checkContainsMap(checker, ovm, ivm) {
					return false
				}
				continue
			}
		}
		if !reflect.DeepEqual(ov, iv) {
			return false
		}
	}
	return true
}
