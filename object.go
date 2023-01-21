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
	noCopy noCopy
	chain  *chain
	value  map[string]interface{}
}

// NewObject returns a new Object instance.
//
// If reporter is nil, the function panics.
// If value is nil, failure is reported.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
func NewObject(reporter Reporter, value map[string]interface{}) *Object {
	return newObject(newChainWithDefaults("Object()", reporter), value)
}

// NewObjectC returns a new Object instance with config.
//
// Requirements for config are same as for WithConfig function.
// If value is nil, failure is reported.
//
// Example:
//
//	object := NewObjectC(config, map[string]interface{}{"foo":123})
func NewObjectC(config Config, value map[string]interface{}) *Object {
	return newObject(newChainWithConfig("Object()", config.withDefaults()), value)
}

func newObject(parent *chain, val map[string]interface{}) *Object {
	o := &Object{chain: parent.clone(), value: nil}

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

// Keys returns a new Array instance with object's keys.
// Keys are sorted in ascending order.
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
	for _, kv := range o.sortedKV() {
		keys = append(keys, kv.key)
	}

	return newArray(o.chain, keys)
}

// Values returns a new Array instance with object's values.
// Values are sorted by keys ascending order.
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

	values := []interface{}{}
	for _, kv := range o.sortedKV() {
		values = append(values, kv.val)
	}

	return newArray(o.chain, values)
}

// Value returns a new Value instance with value for given key.
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

// Iter returns a new map of Values attached to object elements.
//
// Example:
//
//	numbers := map[string]interface{}{"foo": 123, "bar": 456}
//	object := NewObject(t, numbers)
//
//	for key, value := range object.Iter() {
//	    value.Number().Equal(numbers[key])
//	}
func (o *Object) Iter() map[string]Value {
	o.chain.enter("Iter()")
	defer o.chain.leave()

	if o.chain.failed() {
		return map[string]Value{}
	}

	obj := make(map[string]Value)

	for k, v := range o.value {
		valueChain := o.chain.clone()
		valueChain.replace("Iter[%q]", k)

		obj[k] = *newValue(valueChain, v)
	}

	return obj
}

// Every runs the passed function for all the key value pairs in the object.
//
// If assertion inside function fails, the original Object is marked failed.
//
// Every will execute the function for all values in the object irrespective
// of assertion failures for some values in the object.
//
// The function is invoked for key value pairs sorted by keys in ascending order.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123, "bar": 456})
//
//	object.Every(func(key string, value *httpexpect.Value) {
//	  value.String().NotEmpty()
//	})
func (o *Object) Every(fn func(key string, value *Value)) *Object {
	o.chain.enter("Every()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if fn == nil {
		o.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return o
	}

	chainFailure := false

	for _, kv := range o.sortedKV() {
		valueChain := o.chain.clone()
		valueChain.replace("Every[%q]", kv.key)

		valueChain.setFailCallback(func() {
			chainFailure = true
		})

		fn(kv.key, newValue(valueChain, kv.val))
	}

	if chainFailure {
		o.chain.setFailed()
	}

	return o
}

// Filter accepts a function that returns a boolean. The function is ran
// over the object elements. If the function returns true, the element passes
// the filter and is added to the new object of filtered elements. If false,
// the value is skipped (or in other words filtered out). After iterating
// through all the elements of the original object, the new filtered object
// is returned.
//
// If there are any failed assertions in the filtering function, the
// element is omitted without causing test failure.
//
// The function is invoked for key value pairs sorted by keys in ascending order.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{
//		"foo": "bar",
//		"baz": 6,
//		"qux": "quux",
//	})
//	filteredObject := object.Filter(func(key string, value *httpexpect.Value) bool {
//		value.String().NotEmpty()		//fails on 6
//		return value.Raw() != "bar"		//fails on "bar"
//	})
//	filteredObject.Equal(map[string]interface{}{"qux":"quux"})	//succeeds
func (o *Object) Filter(fn func(key string, value *Value) bool) *Object {
	o.chain.enter("Filter()")
	defer o.chain.leave()

	if o.chain.failed() {
		return newObject(o.chain, nil)
	}

	if fn == nil {
		o.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return newObject(o.chain, nil)
	}

	filteredObject := make(map[string]interface{})

	for _, kv := range o.sortedKV() {
		valueChain := o.chain.clone()
		valueChain.replace("Filter[%q]", kv.key)

		valueChain.setSeverity(SeverityLog)

		chainFailed := false
		valueChain.setFailCallback(func() {
			chainFailed = true
		})

		if fn(kv.key, newValue(valueChain, kv.val)) && !chainFailed {
			filteredObject[kv.key] = kv.val
		}
	}

	return newObject(o.chain, filteredObject)
}

// Transform runs the passed function on all the Elements in the Object
// and returns a new object without effecting original object.
//
// The function is invoked for key value pairs sorted by keys in ascending order.
//
// Example:
//
//	object := NewObject(t, []interface{}{"x": "foo", "y": "bar"})
//	transformedObject := object.Transform(
//		func(key string, value interface{}) interface{} {
//			return strings.ToUpper(value.(string))
//		})
//	transformedObject.Equals([]interface{}{"x": "FOO", "y": "BAR"})
func (o *Object) Transform(fn func(key string, value interface{}) interface{}) *Object {
	o.chain.enter("Transform()")
	defer o.chain.leave()

	if o.chain.failed() {
		return newObject(o.chain, nil)
	}

	if fn == nil {
		o.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return newObject(o.chain, nil)
	}

	object := map[string]interface{}{}

	for _, kv := range o.sortedKV() {
		transformedValue := fn(kv.key, kv.val)
		object[kv.key] = transformedValue
	}

	return newObject(o.chain, object)
}

// Find accepts a function that returns a boolean, runs it over the object
// elements, and returns the first element on which it returned true.
//
// If there are any failed assertions in the predicate function, the
// element is skipped without causing test failure.
//
// If no elements were found, a failure is reported.
//
// The function is invoked for key value pairs sorted by keys in ascending order.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{
//		"a": 1,
//		"b": "foo",
//		"c": 101,
//		"d": "bar",
//		"e": 201,
//	})
//	foundValue := object.Find(func(key string, value *httpexpect.Value)  bool {
//		num := value.Number()      // skip if element is not a string
//		return num.Raw() > 100     // check element value
//	})
//	foundValue.Equal(101) // succeeds
func (o *Object) Find(fn func(key string, value *Value) bool) *Value {
	o.chain.enter("Find()")
	defer o.chain.leave()

	if o.chain.failed() {
		return newValue(o.chain, nil)
	}

	if fn == nil {
		o.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return newValue(o.chain, nil)
	}

	for _, kv := range o.sortedKV() {
		valueChain := o.chain.clone()
		valueChain.replace("Find[%q]", kv.key)

		valueChain.setSeverity(SeverityLog)

		chainFailed := false
		valueChain.setFailCallback(func() {
			chainFailed = true
		})

		if fn(kv.key, newValue(valueChain, kv.val)) && !chainFailed {
			return newValue(o.chain, kv.val)
		}
	}

	o.chain.fail(AssertionFailure{
		Type:   AssertValid,
		Actual: &AssertionValue{o.value},
		Errors: []error{
			errors.New("expected: at least one object element matches predicate"),
		},
	})

	return newValue(o.chain, nil)
}

// FindAll accepts a function that returns a boolean, runs it over the object
// elements, and returns all the elements on which it returned true.
//
// If there are any failed assertions in the predicate function, the
// element is skipped without causing test failure.
//
// If no elements were found, empty slice is returned without reporting error.
//
// The function is invoked for key value pairs sorted by keys in ascending order.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{
//		"a": 1,
//		"b": "foo",
//		"c": 101,
//		"d": "bar",
//		"e": 201,
//	})
//	foundValues := object.FindAll(func(key string, value *httpexpect.Value)  bool {
//		num := value.Number()      // skip if element is not a string
//		return num.Raw() > 100     // check element value
//	})
//
//	assert.Equal(t, len(foundValues), 2)
//	foundValues[0].Equal(101)
//	foundValues[1].Equal(201)
func (o *Object) FindAll(fn func(key string, value *Value) bool) []*Value {
	o.chain.enter("FindAll()")
	defer o.chain.leave()

	if o.chain.failed() {
		return []*Value{}
	}

	if fn == nil {
		o.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return []*Value{}
	}

	foundValues := make([]*Value, 0, len(o.value))

	for _, kv := range o.sortedKV() {
		valueChain := o.chain.clone()
		valueChain.replace("FindAll[%q]", kv.key)

		valueChain.setSeverity(SeverityLog)

		chainFailed := false
		valueChain.setFailCallback(func() {
			chainFailed = true
		})

		if fn(kv.key, newValue(valueChain, kv.val)) && !chainFailed {
			foundValues = append(foundValues, newValue(o.chain, kv.val))
		}
	}

	return foundValues
}

// NotFind accepts a function that returns a boolean, runs it over the object
// elelements, and checks that it does not return true for any of the elements.
//
// If there are any failed assertions in the predicate function, the
// element is skipped without causing test failure.
//
// If the predicate function did not fail and returned true for at least
// one element, a failure is reported.
//
// The function is invoked for key value pairs sorted by keys in ascending order.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{
//		"a": 1,
//		"b": "foo",
//		"c": 2,
//		"d": "bar",
//	})
//	object.NotFind(func(key string, value *httpexpect.Value) bool {
//		num := value.Number()    // skip if element is not a number
//		return num.Raw() > 100   // check element value
//	}) // succeeds
func (o *Object) NotFind(fn func(key string, value *Value) bool) *Object {
	o.chain.enter("NotFind()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if fn == nil {
		o.chain.fail(AssertionFailure{
			Type: AssertUsage,
			Errors: []error{
				errors.New("unexpected nil function argument"),
			},
		})
		return o
	}

	for _, kv := range o.sortedKV() {
		valueChain := o.chain.clone()
		valueChain.replace("NotFind[%q]", kv.key)

		valueChain.setSeverity(SeverityLog)

		chainFailed := false
		valueChain.setFailCallback(func() {
			chainFailed = true
		})

		if fn(kv.key, newValue(valueChain, kv.val)) && !chainFailed {
			o.chain.fail(AssertionFailure{
				Type:     AssertNotContainsElement,
				Expected: &AssertionValue{kv.val},
				Actual:   &AssertionValue{o.value},
				Errors: []error{
					errors.New("expected: none of the object elements match predicate"),
					fmt.Errorf("element with key %q matches predicate", kv.key),
				},
			})
			return o
		}
	}

	return o
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

// Equal succeeds if object is equal to given value.
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

// NotEqual succeeds if object is not equal to given value.
// Before comparison, both object and value are converted to canonical form.
//
// value should be map[string]interface{} or struct.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	object.Equal(map[string]interface{}{"bar": 123})
func (o *Object) NotEqual(value interface{}) *Object {
	o.chain.enter("NotEqual()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	expected, ok := canonMap(o.chain, value)
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

// ContainsValue succeeds if object contains given value with any key.
// Before comparison, both object and value are converted to canonical form.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	object.ContainsValue(123)
func (o *Object) ContainsValue(value interface{}) *Object {
	o.chain.enter("ContainsValue()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if _, ok := o.containsValue(value); !ok {
		o.chain.fail(AssertionFailure{
			Type:     AssertContainsElement,
			Actual:   &AssertionValue{o.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: map contains element (with any key)"),
			},
		})
	}

	return o
}

// NotContainsValue succeeds if object does not contain given value with any key.
// Before comparison, both object and value are converted to canonical form.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	object.NotContainsValue(456)
func (o *Object) NotContainsValue(value interface{}) *Object {
	o.chain.enter("NotContainsValue()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if key, ok := o.containsValue(value); ok {
		o.chain.fail(AssertionFailure{
			Type:     AssertNotContainsElement,
			Actual:   &AssertionValue{o.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: map does not contain element (with any key)"),
				fmt.Errorf("found matching element with key %q", key),
			},
		})
	}

	return o
}

// ContainsSubset succeeds if given value is a subset of object.
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
//	object.ContainsSubset(map[string]interface{}{  // success
//	    "foo": 123,
//	    "bar": map[string]interface{}{
//	        "a": true,
//	    },
//	})
//
//	object.ContainsSubset(map[string]interface{}{  // failure
//	    "foo": 123,
//	    "qux": 456,
//	})
//
//	object.ContainsSubset(map[string]interface{}{  // failure, slices should match exactly
//	    "bar": []interface{}{"x"},
//	})
func (o *Object) ContainsSubset(value interface{}) *Object {
	o.chain.enter("ContainsSubset()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if !o.containsSubset(value) {
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

// NotContainsSubset succeeds if given value is not a subset of object.
// Before comparison, both object and value are converted to canonical form.
//
// value should be map[string]interface{} or struct.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123, "bar": 456})
//	object.NotContainsSubset(map[string]interface{}{"foo": 123, "bar": "no-no-no"})
func (o *Object) NotContainsSubset(value interface{}) *Object {
	o.chain.enter("NotContainsSubset()")
	defer o.chain.leave()

	if o.chain.failed() {
		return o
	}

	if o.containsSubset(value) {
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

// Deprecated: use ContainsSubset instead.
func (o *Object) ContainsMap(value interface{}) *Object {
	return o.ContainsSubset(value)
}

// Deprecated: use NotContainsSubset instead.
func (o *Object) NotContainsMap(value interface{}) *Object {
	return o.NotContainsSubset(value)
}

// ValueEqual succeeds if object's value for given key is equal to given value.
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

// NotValueEqual succeeds if object's value for given key is not equal to given
// value. Before comparison, both values are converted to canonical form.
//
// value should be map[string]interface{} or struct.
//
// If object doesn't contain any value for given key, failure is reported.
//
// Example:
//
//	object := NewObject(t, map[string]interface{}{"foo": 123})
//	object.NotValueEqual("foo", "bad value")  // success
//	object.NotValueEqual("bar", "bad value")  // failure! (key is missing)
func (o *Object) NotValueEqual(key string, value interface{}) *Object {
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

// Deprecated: use NotValueEqual instead.
func (o *Object) ValueNotEqual(key string, value interface{}) *Object {
	return o.NotValueEqual(key, value)
}

type kv struct {
	key string
	val interface{}
}

func (o *Object) sortedKV() []kv {
	kvs := make([]kv, 0, len(o.value))
	for key, val := range o.value {
		kvs = append(kvs, kv{key: key, val: val})
	}
	sort.Slice(kvs, func(i, j int) bool { return kvs[i].key < kvs[j].key })
	return kvs
}

func (o *Object) containsKey(arg string) bool {
	for k := range o.value {
		if k == arg {
			return true
		}
	}
	return false
}

func (o *Object) containsValue(arg interface{}) (string, bool) {
	value, ok := canonValue(o.chain, arg)
	if !ok {
		return "", false
	}
	for k, v := range o.value {
		if reflect.DeepEqual(value, v) {
			return k, true
		}
	}
	return "", false
}

func (o *Object) containsSubset(arg interface{}) bool {
	value, ok := canonMap(o.chain, arg)
	if !ok {
		return false
	}
	return checkSubset(o.value, value)
}

func checkSubset(outer, inner map[string]interface{}) bool {
	for k, iv := range inner {
		ov, ok := outer[k]
		if !ok {
			return false
		}
		if ovm, ok := ov.(map[string]interface{}); ok {
			if ivm, ok := iv.(map[string]interface{}); ok {
				if !checkSubset(ovm, ivm) {
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
