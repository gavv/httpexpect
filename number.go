package httpexpect

type Number struct {
	checker Checker
	value   float64
}

func NewNumber(checker Checker, value float64) *Number {
	return &Number{checker, value}
}

func (n *Number) Raw() float64 {
	return n.value
}

func (n *Number) Equal(expected interface{}) *Number {
	v, ok := canonNumber(n.checker, expected)
	if !ok {
		return n
	}
	if !(n.value == v) {
		n.checker.Fail("expected number == %v, got %v", v, n.value)
	}
	return n
}

func (n *Number) NotEqual(expected interface{}) *Number {
	v, ok := canonNumber(n.checker, expected)
	if !ok {
		return n
	}
	if !(n.value != v) {
		n.checker.Fail("expected number != %v, got %v", v, n.value)
	}
	return n
}

func (n *Number) Gt(expected interface{}) *Number {
	v, ok := canonNumber(n.checker, expected)
	if !ok {
		return n
	}
	if !(n.value > v) {
		n.checker.Fail("expected number > %v, got %v", v, n.value)
	}
	return n
}

func (n *Number) Ge(expected interface{}) *Number {
	v, ok := canonNumber(n.checker, expected)
	if !ok {
		return n
	}
	if !(n.value >= v) {
		n.checker.Fail("expected number >= %v, got %v", v, n.value)
	}
	return n
}

func (n *Number) Lt(expected interface{}) *Number {
	v, ok := canonNumber(n.checker, expected)
	if !ok {
		return n
	}
	if !(n.value < v) {
		n.checker.Fail("expected number < %v, got %v", v, n.value)
	}
	return n
}

func (n *Number) Le(expected interface{}) *Number {
	v, ok := canonNumber(n.checker, expected)
	if !ok {
		return n
	}
	if !(n.value <= v) {
		n.checker.Fail("expected number <= %v, got %v", v, n.value)
	}
	return n
}

func (n *Number) InRange(ra, rb interface{}) *Number {
	a, ok := canonNumber(n.checker, ra)
	if !ok {
		return n
	}
	b, ok := canonNumber(n.checker, rb)
	if !ok {
		return n
	}
	if !(n.value >= a && n.value <= b) {
		n.checker.Fail("expected number in range [%v; %v], got %v", a, b, n.value)
	}
	return n
}
