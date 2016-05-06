package httpexpect

type Array struct {
	checker Checker
	value   []interface{}
}

func NewArray(checker Checker, value []interface{}) *Array {
	if value == nil {
		checker.Fail("expected non-nil array value")
	} else {
		value, _ = canonArray(checker, value)
	}
	return &Array{checker, value}
}

func (a *Array) Raw() []interface{} {
	return a.value
}

func (a *Array) Length() *Number {
	return NewNumber(a.checker.Clone(), float64(len(a.value)))
}

func (a *Array) Element(index int) *Value {
	if len(a.value) <= index {
		a.checker.Fail("expected array length > %d, got %d", index, len(a.value))
		return NewValue(a.checker.Clone(), nil)
	}
	return NewValue(a.checker.Clone(), a.value[index])
}

func (a *Array) Empty() *Array {
	expected := make([]interface{}, 0)
	a.checker.Equal(expected, a.value)
	return a
}

func (a *Array) NotEmpty() *Array {
	expected := make([]interface{}, 0)
	a.checker.NotEqual(expected, a.value)
	return a
}

func (a *Array) Equal(v []interface{}) *Array {
	expected, ok := canonArray(a.checker, v)
	if !ok {
		return a
	}
	a.checker.Equal(expected, a.value)
	return a
}

func (a *Array) NotEqual(v []interface{}) *Array {
	expected, ok := canonArray(a.checker, v)
	if !ok {
		return a
	}
	a.checker.NotEqual(expected, a.value)
	return a
}

func (a *Array) Contains(v... interface{}) *Array {
	elements, ok := canonArray(a.checker, v)
	if !ok {
		return a
	}
	for _, e := range elements {
		if !a.containsElement(e) {
			a.checker.Fail("expected array containing %v, got %v", e, a.value)
		}
	}
	return a
}

func (a *Array) NotContains(v... interface{}) *Array {
	elements, ok := canonArray(a.checker, v)
	if !ok {
		return a
	}
	for _, e := range elements {
		if a.containsElement(e) {
			a.checker.Fail("expected array NOT containing %v, got %v", e, a.value)
		}
	}
	return a
}

func (a *Array) Elements(v... interface{}) *Array {
	return a.Equal(v)
}

func (a *Array) ElementsAnyOrder(v... interface{}) *Array {
	elements, ok := canonArray(a.checker, v)
	if !ok {
		return a
	}
	if len(elements) != len(a.value) {
		a.checker.Fail("expected array len == %d, got %d", len(elements), len(a.value))
		return a
	}
	for _, e := range elements {
		if !a.containsElement(e) {
			a.checker.Fail("expected array containing %v, got %v", e, a.value)
		}
	}
	return a
}

func (a *Array) containsElement(expected interface{}) bool {
	for _, e := range a.value {
		if a.checker.Compare(expected, e) {
			return true
		}
	}
	return false
}
