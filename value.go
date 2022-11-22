package httpexpect

import (
	"errors"
	"reflect"
)

// Value provides methods to inspect attached interface{} object
// (Go representation of arbitrary JSON value) and cast it to
// concrete type.
type Value struct {
	chain *chain
	value interface{}
}

// NewValue returns a new Value instance.
//
// reporter should not be nil, but value may be nil.
//
// Example:
//
//	value := NewValue(t, map[string]interface{}{"foo": 123})
//	value.Object()
//
//	value := NewValue(t, []interface{}{"foo", 123})
//	value.Array()
//
//	value := NewValue(t, "foo")
//	value.String()
//
//	value := NewValue(t, 123)
//	value.Number()
//
//	value := NewValue(t, true)
//	value.Boolean()
//
//	value := NewValue(t, nil)
//	value.Null()
func NewValue(reporter Reporter, value interface{}) *Value {
	return newValue(newChainWithDefaults("Value()", reporter), value)
}

func newValue(parent *chain, val interface{}) *Value {
	v := &Value{parent.clone(), nil}

	if val != nil {
		v.value, _ = canonValue(v.chain, val)
	}

	return v
}

// Raw returns underlying value attached to Value.
// This is the value originally passed to NewValue, converted to canonical form.
//
// Example:
//
//	value := NewValue(t, "foo")
//	assert.Equal(t, "foo", number.Raw().(string))
func (v *Value) Raw() interface{} {
	return v.value
}

// Path returns a new Value object for child object(s) matching given
// JSONPath expression.
//
// JSONPath is a simple XPath-like query language.
// See http://goessner.net/articles/JsonPath/.
//
// We currently use https://github.com/yalp/jsonpath, which implements
// only a subset of JSONPath, yet useful for simple queries. It doesn't
// support filters and requires double quotes for strings.
//
// Example 1:
//
//	json := `{"users": [{"name": "john"}, {"name": "bob"}]}`
//	value := NewValue(t, json)
//
//	value.Path("$.users[0].name").String().Equal("john")
//	value.Path("$.users[1].name").String().Equal("bob")
//
// Example 2:
//
//	json := `{"yfGH2a": {"user": "john"}, "f7GsDd": {"user": "john"}}`
//	value := NewValue(t, json)
//
//	for _, user := range value.Path("$..user").Array().Iter() {
//	    user.String().Equal("john")
//	}
func (v *Value) Path(path string) *Value {
	v.chain.enter("Path(%q)", path)
	defer v.chain.leave()

	return jsonPath(v.chain, v.value, path)
}

// Schema succeeds if value matches given JSON Schema.
//
// JSON Schema specifies a JSON-based format to define the structure of
// JSON data. See http://json-schema.org/.
// We use https://github.com/xeipuuv/gojsonschema implementation.
//
// schema should be one of the following:
//   - go value that can be json.Marshal-ed to a valid schema
//   - type convertible to string containing valid schema
//   - type convertible to string containing valid http:// or file:// URI,
//     pointing to reachable and valid schema
//
// Example 1:
//
//	 schema := `{
//	   "type": "object",
//	   "properties": {
//	      "foo": {
//	          "type": "string"
//	      },
//	      "bar": {
//	          "type": "integer"
//	      }
//	  },
//	  "require": ["foo", "bar"]
//	}`
//
//	value := NewValue(t, map[string]interface{}{
//	    "foo": "a",
//	    "bar": 1,
//	})
//
//	value.Schema(schema)
//
// Example 2:
//
//	value := NewValue(t, data)
//	value.Schema("http://example.com/schema.json")
func (v *Value) Schema(schema interface{}) *Value {
	v.chain.enter("Schema()")
	defer v.chain.leave()

	jsonSchema(v.chain, v.value, schema)
	return v
}

// Object returns a new Object attached to underlying value.
//
// If underlying value is not an object (map[string]interface{}), failure is reported
// and empty (but non-nil) value is returned.
//
// Example:
//
//	value := NewValue(t, map[string]interface{}{"foo": 123})
//	value.Object().ContainsKey("foo")
func (v *Value) Object() *Object {
	v.chain.enter("Object()")
	defer v.chain.leave()

	if v.chain.failed() {
		return newObject(v.chain, nil)
	}

	data, ok := v.value.(map[string]interface{})

	if !ok {
		v.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{v.value},
			Errors: []error{
				errors.New("expected: value is map"),
			},
		})
		return newObject(v.chain, nil)
	}

	return newObject(v.chain, data)
}

// Array returns a new Array attached to underlying value.
//
// If underlying value is not an array ([]interface{}), failure is reported and empty
// (but non-nil) value is returned.
//
// Example:
//
//	value := NewValue(t, []interface{}{"foo", 123})
//	value.Array().Elements("foo", 123)
func (v *Value) Array() *Array {
	v.chain.enter("Array()")
	defer v.chain.leave()

	if v.chain.failed() {
		return newArray(v.chain, nil)
	}

	data, ok := v.value.([]interface{})

	if !ok {
		v.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{v.value},
			Errors: []error{
				errors.New("expected: value is array"),
			},
		})
		return newArray(v.chain, nil)
	}

	return newArray(v.chain, data)
}

// String returns a new String attached to underlying value.
//
// If underlying value is not a string, failure is reported and empty (but non-nil)
// value is returned.
//
// Example:
//
//	value := NewValue(t, "foo")
//	value.String().EqualFold("FOO")
func (v *Value) String() *String {
	v.chain.enter("String()")
	defer v.chain.leave()

	if v.chain.failed() {
		return newString(v.chain, "")
	}

	data, ok := v.value.(string)

	if !ok {
		v.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{v.value},
			Errors: []error{
				errors.New("expected: value is string"),
			},
		})
		return newString(v.chain, "")
	}

	return newString(v.chain, data)
}

// Number returns a new Number attached to underlying value.
//
// If underlying value is not a number (numeric type convertible to float64), failure
// is reported and empty (but non-nil) value is returned.
//
// Example:
//
//	value := NewValue(t, 123)
//	value.Number().InRange(100, 200)
func (v *Value) Number() *Number {
	v.chain.enter("Number()")
	defer v.chain.leave()

	if v.chain.failed() {
		return newNumber(v.chain, 0)
	}

	data, ok := v.value.(float64)

	if !ok {
		v.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{v.value},
			Errors: []error{
				errors.New("expected: value is number"),
			},
		})
		return newNumber(v.chain, 0)
	}

	return newNumber(v.chain, data)
}

// Boolean returns a new Boolean attached to underlying value.
//
// If underlying value is not a bool, failure is reported and empty (but non-nil)
// value is returned.
//
// Example:
//
//	value := NewValue(t, true)
//	value.Boolean().True()
func (v *Value) Boolean() *Boolean {
	v.chain.enter("Boolean()")
	defer v.chain.leave()

	if v.chain.failed() {
		return newBoolean(v.chain, false)
	}

	data, ok := v.value.(bool)

	if !ok {
		v.chain.fail(AssertionFailure{
			Type:   AssertValid,
			Actual: &AssertionValue{v.value},
			Errors: []error{
				errors.New("expected: value is boolean"),
			},
		})
		return newBoolean(v.chain, false)
	}

	return newBoolean(v.chain, data)
}

// Null succeeds if value is nil.
//
// Note that non-nil interface{} that points to nil value (e.g. nil slice or map)
// is also treated as null value. Empty (non-nil) slice or map, empty string, and
// zero number are not treated as null value.
//
// Example:
//
//	value := NewValue(t, nil)
//	value.Null()
//
//	value := NewValue(t, []interface{}(nil))
//	value.Null()
func (v *Value) Null() *Value {
	v.chain.enter("Null()")
	defer v.chain.leave()

	if v.chain.failed() {
		return v
	}

	if !(v.value == nil) {
		v.chain.fail(AssertionFailure{
			Type:   AssertNil,
			Actual: &AssertionValue{v.value},
			Errors: []error{
				errors.New("expected: value is null"),
			},
		})
	}

	return v
}

// NotNull succeeds if value is not nil.
//
// Note that non-nil interface{} that points to nil value (e.g. nil slice or map)
// is also treated as null value. Empty (non-nil) slice or map, empty string, and
// zero number are not treated as null value.
//
// Example:
//
//	value := NewValue(t, "")
//	value.NotNull()
//
//	value := NewValue(t, make([]interface{}, 0)
//	value.Null()
func (v *Value) NotNull() *Value {
	v.chain.enter("NotNull()")
	defer v.chain.leave()

	if v.chain.failed() {
		return v
	}

	if !(v.value != nil) {
		v.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{v.value},
			Errors: []error{
				errors.New("expected: value is non-null"),
			},
		})
	}

	return v
}

// Equal succeeds if value is equal to another value (e.g. map, slice, string, etc).
// Before comparison, both values are converted to canonical form.
//
// Example:
//
//	value := NewValue(t, "foo")
//	value.Equal("foo")
func (v *Value) Equal(value interface{}) *Value {
	v.chain.enter("Equal()")
	defer v.chain.leave()

	if v.chain.failed() {
		return v
	}

	expected, ok := canonValue(v.chain, value)
	if !ok {
		return v
	}

	if !reflect.DeepEqual(expected, v.value) {
		v.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{v.value},
			Expected: &AssertionValue{expected},
			Errors: []error{
				errors.New("expected: values are equal"),
			},
		})
	}

	return v
}

// NotEqual succeeds if value is not equal to another value (e.g. map, slice,
// string, etc). Before comparison, both values are converted to canonical form.
//
// Example:
//
//	value := NewValue(t, "foo")
//	value.NorEqual("bar")
func (v *Value) NotEqual(value interface{}) *Value {
	v.chain.enter("NotEqual()")
	defer v.chain.leave()

	if v.chain.failed() {
		return v
	}

	expected, ok := canonValue(v.chain, value)
	if !ok {
		return v
	}

	if reflect.DeepEqual(expected, v.value) {
		v.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{v.value},
			Expected: &AssertionValue{expected},
			Errors: []error{
				errors.New("expected: values are non-equal"),
			},
		})
	}

	return v
}
