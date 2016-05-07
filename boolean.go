package httpexpect

// Boolean provides methods to inspect attached bool value
// (Go representation of JSON boolean).
type Boolean struct {
	checker Checker
	value   bool
}

// NewBoolean returns a new Boolean given a checker used to report failures
// and value to be inspected.
//
// checker should not be nil.
//
// Example:
//  boolean := NewBoolean(NewAssertChecker(t), true)
func NewBoolean(checker Checker, value bool) *Boolean {
	return &Boolean{checker, value}
}

// Raw returns underlying value attached to Boolean.
// This is the value originally passed to NewBoolean.
//
// Example:
//  boolean := NewBoolean(checker, true)
//  assert.Equal(t, true, boolean.Raw())
func (b *Boolean) Raw() bool {
	return b.value
}

// Equal succeedes if boolean is equal to given value.
//
// Example:
//  boolean := NewBoolean(checker, true)
//  boolean.Equal(true)
func (b *Boolean) Equal(v bool) *Boolean {
	if !(b.value == v) {
		b.checker.Fail("expected boolean == %v, got %v", v, b.value)
	}
	return b
}

// NotEqual succeedes if boolean is not equal to given value.
//
// Example:
//  boolean := NewBoolean(checker, true)
//  boolean.NotEqual(false)
func (b *Boolean) NotEqual(v bool) *Boolean {
	if !(b.value != v) {
		b.checker.Fail("expected boolean != %v, got %v", v, b.value)
	}
	return b
}

// True succeedes if boolean is true.
//
// Example:
//  boolean := NewBoolean(checker, true)
//  boolean.True()
func (b *Boolean) True() *Boolean {
	return b.Equal(true)
}

// True succeedes if boolean is false.
//
// Example:
//  boolean := NewBoolean(checker, false)
//  boolean.False()
func (b *Boolean) False() *Boolean {
	return b.Equal(false)
}
