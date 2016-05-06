package httpexpect

type Object struct {
	checker Checker
	value   map[string]interface{}
	canon   map[string]interface{}
}

func NewObject(checker Checker, value map[string]interface{}) *Object {
	o := &Object{checker, value, value}
	v, ok := canonMap(o.checker, o.value)
	if ok {
		o.canon = v
	}
	return o
}

func (o *Object) Raw() map[string]interface{} {
	return o.value
}

func (o *Object) Keys() *Array {
	keys := make([]interface{}, 0)
	for k, _ := range o.value {
		keys = append(keys, k)
	}
	return NewArray(o.checker.Clone(), keys)
}

func (o *Object) Values() *Array {
	values := make([]interface{}, 0)
	for _, v := range o.value {
		values = append(values, v)
	}
	return NewArray(o.checker.Clone(), values)
}

func (o *Object) Value(key string) *Value {
	value, ok := o.value[key]
	if !ok {
		o.checker.Fail("expected map containing %v, got %v", key, o.value)
		return NewValue(o.checker.Clone(), nil)
	}
	return NewValue(o.checker.Clone(), value)
}

func (o *Object) Empty() *Object {
	expected := make(map[string]interface{})
	actual := o.canon
	if actual == nil {
		actual = make(map[string]interface{})
	}
	o.checker.Equal(expected, actual)
	return o
}

func (o *Object) NotEmpty() *Object {
	expected := make(map[string]interface{})
	actual := o.canon
	if actual == nil {
		actual = make(map[string]interface{})
	}
	o.checker.NotEqual(expected, actual)
	return o
}

func (o *Object) Equal(v map[string]interface{}) *Object {
	expected, ok := canonMap(o.checker, v)
	if !ok {
		return o
	}
	o.checker.Equal(expected, o.canon)
	return o
}

func (o *Object) NotEqual(v map[string]interface{}) *Object {
	expected, ok := canonMap(o.checker, v)
	if !ok {
		return o
	}
	o.checker.NotEqual(expected, o.canon)
	return o
}

func (o *Object) ContainsKey(key string) *Object {
	if !o.containsKey(key) {
		o.checker.Fail("expected map containing %v, got %v", key, o.value)
	}
	return o
}

func (o *Object) NotContainsKey(key string) *Object {
	if o.containsKey(key) {
		o.checker.Fail("expected map NOT containing %v, got %v", key, o.value)
	}
	return o
}

func (o *Object) ContainsMap(submap map[string]interface{}) *Object {
	if !o.containsMap(submap) {
		o.checker.Fail("expected map containing submap %v, got %v", submap, o.value)
	}
	return o
}

func (o *Object) NotContainsMap(submap map[string]interface{}) *Object {
	if o.containsMap(submap) {
		o.checker.Fail("expected map NOT containing submap %v, got %v", submap, o.value)
	}
	return o
}

func (o *Object) ValueEqual(k string, v interface{}) *Object {
	if !o.containsKey(k) {
		o.checker.Fail("expected map containing %v, got %v", k, o.value)
		return o
	}
	expected, ok := canonValue(o.checker, v)
	if !ok {
		return o
	}
	actual, ok := canonValue(o.checker, o.canon[k])
	if !ok {
		return o
	}
	o.checker.Equal(expected, actual)
	return o
}

func (o *Object) ValueNotEqual(k string, v interface{}) *Object {
	if !o.containsKey(k) {
		o.checker.Fail("expected map containing %v, got %v", k, o.value)
		return o
	}
	expected, ok := canonValue(o.checker, v)
	if !ok {
		return o
	}
	o.checker.NotEqual(expected, o.canon[k])
	return o
}

func (o *Object) containsKey(key string) bool {
	for k, _ := range o.canon {
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
		if !o.checker.Compare(v, o.canon[k]) {
			return false
		}
	}
	return true
}
