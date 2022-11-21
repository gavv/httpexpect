package httpexpect

import (
	"errors"
	"fmt"
	"reflect"
	"sort"
)

// Object provides methods to inspect attached map[string]interface{} object
// (Go representation of JSON object).
type Object struct {
	chain *chain
	value map[string]interface{}
}

// NewObject returns a new Object given a reporter used to report failures
// and value to be inspected.
//
// Both reporter and value should not be nil. If value is nil, failure is
// reported.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
func NewObject(reporter Reporter, value map[string]interface{}) *Object {
	return newObject(newChainWithDefaults("Object()", reporter), value)
}

func newObject(parent *chain, val map[string]interface{}) *Object {
	o := &Object{parent.clone(), nil}

	if val == nil {
		o.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{val},
			Errors: []error{
				errors.New("expected: non-nil map"),
			},
		})
	} else {
		o.value, _ = canonMap(o.chain, val)
	}

	return o
}

// Raw returns underlying value attached to Object.
// This is the value originally passed to NewObject, converted to canonical form.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	assert.Equal(t, map[string]interface{}{"foo": 123.0}, object.Raw())
func (o *Object) Raw() map[string]interface{} {
	return o.value
}

// Path is similar to Value.Path.
func (o *Object) Path(path string) *Value {
	o.chain.enter("Path(%q)", path)
	defer o.chain.leave()

	return jsonPath(o.chain, o.value, path)
}

// Schema is similar to Value.Schema.
func (o *Object) Schema(schema interface{}) *Object {
	o.chain.enter("Schema()")
	defer o.chain.leave()

	jsonSchema(o.chain, o.value, schema)
	return o
}

// Keys returns a new Array object that may be used to inspect objects keys.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123, "bar": 456})
//	object.Keys().ContainsOnly("foo", "bar")
func (o *Object) Keys() *Array {
	o.chain.enter("Keys()")
	defer o.chain.leave()

	if o.chain.failed() {
		return newArray(o.chain, nil)
	}

	keys := []interface{}{}
	for k := range o.value {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i].(string) < keys[j].(string)
	})

	return newArray(o.chain, keys)
}

// Values returns a new Array object that may be used to inspect objects values.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123, "bar": 456})
//	object.Values().ContainsOnly(123, 456)
func (o *Object) Values() *Array {
	o.chain.enter("Values()")
	defer o.chain.leave()

	if o.chain.failed() {
		return newArray(o.chain, nil)
	}

	keys := []string{}
	for k := range o.value {
		keys = append(keys, k)
	}

	sort.Slice(keys, func(i, j int) bool {
		return keys[i] < keys[j]
	})

	values := []interface{}{}
	for _, k := range keys {
		values = append(values, o.value[k])
	}

	return newArray(o.chain, values)
}

// Value returns a new Value object that may be used to inspect single value
// for given key.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	object.Value("foo").Number().Equal(123)
func (o *Object) Value(key string) *Value {
	o.chain.enter("Value(%q)", key)
	defer o.chain.leave()

	if o.chain.failed() {
		return newValue(o.chain, nil)
	}

	value, ok := o.value[key]

	if !ok {
		o.chain.fail(AssertionFailure{
			Type:     AssertContainsKey,
			Actual:   &AssertionValue{o.value},
			Expected: &AssertionValue{key},
			Errors: []error{
				errors.New("expected: map contains key"),
			},
		})
		return newValue(o.chain, nil)
	}

	return newValue(o.chain, value)
}

// Empty succeeds if object is empty.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{})
//	object.Empty()
func (o *Object) Empty() *Object {
	o.chain.enter("Empty()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if !(len(o.value) == 0) {
		o.chain.fail(AssertionFailure{
			Type:   AssertEmpty,
			Actual: &AssertionValue{o.value},
			Errors: []error{
				errors.New("expected: map is empty"),
			},
		})
	}

	return o
}

// NotEmpty succeeds if object is non-empty.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	object.NotEmpty()
func (o *Object) NotEmpty() *Object {
	o.chain.enter("NotEmpty()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if !(len(o.value) != 0) {
		o.chain.fail(AssertionFailure{
			Type:   AssertNotEmpty,
			Actual: &AssertionValue{o.value},
			Errors: []error{
				errors.New("expected: map is non-empty"),
			},
		})
	}

	return o
}

// Equal succeeds if object is equal to given Go map or struct.
// Before comparison, both object and value are converted to canonical form.
//
// value should be map[string]interface{} or struct.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	object.Equal(map[string]interface{}{"foo": 123})
func (o *Object) Equal(value interface{}) *Object {
	o.chain.enter("Equal()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	expected, ok := canonMap(o.chain, value)
	if !ok {
		return o
	}

	if !reflect.DeepEqual(expected, o.value) {
		o.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{o.value},
			Expected: &AssertionValue{expected},
			Errors: []error{
				errors.New("expected: maps are equal"),
			},
		})
	}

	return o
}

// NotEqual succeeds if object is not equal to given Go map or struct.
// Before comparison, both object and value are converted to canonical form.
//
// value should be map[string]interface{} or struct.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	object.Equal(map[string]interface{}{"bar": 123})
func (o *Object) NotEqual(v interface{}) *Object {
	o.chain.enter("NotEqual()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	expected, ok := canonMap(o.chain, v)
	if !ok {
		return o
	}

	if reflect.DeepEqual(expected, o.value) {
		o.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{o.value},
			Expected: &AssertionValue{expected},
			Errors: []error{
				errors.New("expected: maps are non-equal"),
			},
		})
	}

	return o
}

// ContainsKey succeeds if object contains given key.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	object.ContainsKey("foo")
func (o *Object) ContainsKey(key string) *Object {
	o.chain.enter("ContainsKey()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if !o.containsKey(key) {
		o.chain.fail(AssertionFailure{
			Type:     AssertContainsKey,
			Actual:   &AssertionValue{o.value},
			Expected: &AssertionValue{key},
			Errors: []error{
				errors.New("expected: map contains key"),
			},
		})
	}

	return o
}

// NotContainsKey succeeds if object doesn't contain given key.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	object.NotContainsKey("bar")
func (o *Object) NotContainsKey(key string) *Object {
	o.chain.enter("NotContainsKey()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if o.containsKey(key) {
		o.chain.fail(AssertionFailure{
			Type:     AssertNotContainsKey,
			Actual:   &AssertionValue{o.value},
			Expected: &AssertionValue{key},
			Errors: []error{
				errors.New("expected: map does not contain key"),
			},
		})
	}

	return o
}

// ContainsMap succeeds if object contains given Go value.
// Before comparison, both object and value are converted to canonical form.
//
// value should be map[string]interface{} or struct.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{
//	    "foo": 123,
//	    "bar": []interface{}{"x", "y"},
//	    "bar": map[string]interface{}{
//	        "a": true,
//	        "b": false,
//	    },
//	})
//
//	object.ContainsMap(map[string]interface{}{  // success
//	    "foo": 123,
//	    "bar": map[string]interface{}{
//	        "a": true,
//	    },
//	})
//
//	object.ContainsMap(map[string]interface{}{  // failure
//	    "foo": 123,
//	    "qux": 456,
//	})
//
//	object.ContainsMap(map[string]interface{}{  // failure, slices should match exactly
//	    "bar": []interface{}{"x"},
//	})
func (o *Object) ContainsMap(value interface{}) *Object {
	o.chain.enter("ContainsMap()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if !o.containsMap(value) {
		o.chain.fail(AssertionFailure{
			Type:     AssertContainsSubset,
			Actual:   &AssertionValue{o.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: map contains sub-map"),
			},
		})
	}

	return o
}

// NotContainsMap succeeds if object doesn't contain given Go value.
// Before comparison, both object and value are converted to canonical form.
//
// value should be map[string]interface{} or struct.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123, "bar": 456})
//	object.NotContainsMap(map[string]interface{}{"foo": 123, "bar": "no-no-no"})
func (o *Object) NotContainsMap(value interface{}) *Object {
	o.chain.enter("NotContainsMap()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if o.containsMap(value) {
		o.chain.fail(AssertionFailure{
			Type:     AssertNotContainsSubset,
			Actual:   &AssertionValue{o.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: map does not contain sub-map"),
			},
		})
	}

	return o
}

// ValueEqual succeeds if object's value for given key is equal to given Go value.
// Before comparison, both values are converted to canonical form.
//
// value should be map[string]interface{} or struct.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	object.ValueEqual("foo", 123)
func (o *Object) ValueEqual(key string, value interface{}) *Object {
	o.chain.enter("ValueEqual(%q)", key)
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if !o.containsKey(key) {
		o.chain.fail(AssertionFailure{
			Type:     AssertContainsKey,
			Actual:   &AssertionValue{o.value},
			Expected: &AssertionValue{key},
			Errors: []error{
				errors.New("expected: map contains key"),
			},
		})
		return o
	}

	expected, ok := canonValue(o.chain, value)
	if !ok {
		return o
	}

	if !reflect.DeepEqual(expected, o.value[key]) {
		o.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{o.value[key]},
			Expected: &AssertionValue{value},
			Errors: []error{
				fmt.Errorf(
					"expected: map value for key %q is equal to given value",
					key),
			},
		})
		return o
	}

	return o
}

// ValueNotEqual succeeds if object's value for given key is not equal to given
// Go value. Before comparison, both values are converted to canonical form.
//
// value should be map[string]interface{} or struct.
//
// If object doesn't contain any value for given key, failure is reported.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	object.ValueNotEqual("foo", "bad value")  // success
//	object.ValueNotEqual("bar", "bad value")  // failure! (key is missing)
func (o *Object) ValueNotEqual(key string, value interface{}) *Object {
	o.chain.enter("ValueNotEqual(%q)", key)
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if !o.containsKey(key) {
		o.chain.fail(AssertionFailure{
			Type:     AssertContainsKey,
			Actual:   &AssertionValue{o.value},
			Expected: &AssertionValue{key},
			Errors: []error{
				errors.New("expected: map contains key"),
			},
		})
		return o
	}

	expected, ok := canonValue(o.chain, value)
	if !ok {
		return o
	}

	if reflect.DeepEqual(expected, o.value[key]) {
		o.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{o.value[key]},
			Expected: &AssertionValue{value},
			Errors: []error{
				fmt.Errorf(
					"expected: map value for key %q is non-equal to given value",
					key),
			},
		})
		return o
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
	submap, ok := canonMap(o.chain, sm)
	if !ok {
		return false
	}
	return checkContainsMap(o.value, submap)
}

func checkContainsMap(outer, inner map[string]interface{}) bool {
	for k, iv := range inner {
		ov, ok := outer[k]
		if !ok {
			return false
		}
		if ovm, ok := ov.(map[string]interface{}); ok {
			if ivm, ok := iv.(map[string]interface{}); ok {
				if !checkContainsMap(ovm, ivm) {
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
