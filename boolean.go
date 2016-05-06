package httpexpect

type Boolean struct {
	checker Checker
	value   bool
}

func NewBoolean(checker Checker, value bool) *Boolean {
	return &Boolean{checker, value}
}

func (b *Boolean) Raw() bool {
	return b.value
}

func (b *Boolean) Equal(v bool) *Boolean {
	if !(b.value == v) {
		b.checker.Fail("expected boolean == %v, got %v", v, b.value)
	}
	return b
}

func (b *Boolean) NotEqual(v bool) *Boolean {
	if !(b.value != v) {
		b.checker.Fail("expected boolean != %v, got %v", v, b.value)
	}
	return b
}

func (b *Boolean) True() *Boolean {
	return b.Equal(true)
}

func (b *Boolean) False() *Boolean {
	return b.Equal(false)
}
