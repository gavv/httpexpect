package httpexpect

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
//  object.Keys().ElementsAnyOrder("foo", "bar")
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
//  object.Values().ElementsAnyOrder(123, 456)
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
		o.checker.Fail("expected map containing '%v' key, got %v", key, o.value)
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
	expected := make(map[string]interface{})
	o.checker.Equal(expected, o.value)
	return o
}

// NotEmpty succeedes if object is non-empty.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.NotEmpty()
func (o *Object) NotEmpty() *Object {
	expected := make(map[string]interface{})
	o.checker.NotEqual(expected, o.value)
	return o
}

// Equal succeedes if object is equal to another object.
// Before comparison, both objects are converted to canonical form.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.Equal(map[string]interface{}{"foo": 123})
func (o *Object) Equal(v map[string]interface{}) *Object {
	expected, ok := canonMap(o.checker, v)
	if !ok {
		return o
	}
	o.checker.Equal(expected, o.value)
	return o
}

// NotEqual succeedes if object is not equal to another object.
// Before comparison, both objects are converted to canonical form.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.Equal(map[string]interface{}{"bar": 123})
func (o *Object) NotEqual(v map[string]interface{}) *Object {
	expected, ok := canonMap(o.checker, v)
	if !ok {
		return o
	}
	o.checker.NotEqual(expected, o.value)
	return o
}

// ContainsKey succeedes if object contains given key.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.ContainsKey("foo")
func (o *Object) ContainsKey(key string) *Object {
	if !o.containsKey(key) {
		o.checker.Fail("expected map containing '%v' key, got %v", key, o.value)
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
		o.checker.Fail("expected map NOT containing '%v' key, got %v", key, o.value)
	}
	return o
}

// ContainsMap succeedes if object contains given "subobject".
// Before comparison, both objects are converted to canonical form.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123, "bar": 456})
//  object.ContainsMap(map[string]interface{}{"foo": 123})
//
// This calls are equivalent:
//  object.ContainsMap(m)
//
//  // is equivalent to...
//  for k, v := range m {
//      object.ContainsKey(k).ValueEqual(k, v)
//  }
func (o *Object) ContainsMap(submap map[string]interface{}) *Object {
	if !o.containsMap(submap) {
		o.checker.Fail("expected map containing submap %v, got %v", submap, o.value)
	}
	return o
}

// NotContainsMap succeedes if object doesn't contain given "subobject" exactly.
// Before comparison, both objects are converted to canonical form.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123, "bar": 456})
//  object.NotContainsMap(map[string]interface{}{"foo": 123, "bar": "no-no-no"})
func (o *Object) NotContainsMap(submap map[string]interface{}) *Object {
	if o.containsMap(submap) {
		o.checker.Fail("expected map NOT containing submap %v, got %v", submap, o.value)
	}
	return o
}

// ValueEqual succeedes if object's value for given key is equal to given value.
// Before comparison, both values are converted to canonical form.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.ValueEqual("foo", 123)
func (o *Object) ValueEqual(k string, v interface{}) *Object {
	if !o.containsKey(k) {
		o.checker.Fail("expected map containing '%v' key, got %v", k, o.value)
		return o
	}
	expected, ok := canonValue(o.checker, v)
	if !ok {
		return o
	}
	o.checker.Equal(expected, o.value[k])
	return o
}

// ValueNotEqual succeedes if object's value for given key is not equal to given value.
// Before comparison, both values are converted to canonical form.
//
// If object doesn't contain any value for given key, failure is reported.
//
// Example:
//  object := NewObject(checker, map[string]interface{}{"foo": 123})
//  object.ValueNotEqual("foo", "bad value")  // success
//  object.ValueNotEqual("bar", "bad value")  // failure! (key is missing)
func (o *Object) ValueNotEqual(k string, v interface{}) *Object {
	if !o.containsKey(k) {
		o.checker.Fail("expected map containing '%v' key, got %v", k, o.value)
		return o
	}
	expected, ok := canonValue(o.checker, v)
	if !ok {
		return o
	}
	o.checker.NotEqual(expected, o.value[k])
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

func (o *Object) containsMap(sm map[string]interface{}) bool {
	submap, ok := canonMap(o.checker, sm)
	if !ok {
		return false
	}
	for k, v := range submap {
		if !o.containsKey(k) {
			return false
		}
		if !o.checker.Compare(v, o.value[k]) {
			return false
		}
	}
	return true
}
