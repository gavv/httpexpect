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

// TODO:
//  Keys()
//  Values()
//  Value()

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

func (o *Object) containsKey(key string) bool {
	for k, _ := range o.canon {
		if k == key {
			return true
		}
	}
	return false
}

// TODO:
//  ContainsMap()
//  NotContainsMap()
//  ValueEqual()
//  ValueNotEqual()
