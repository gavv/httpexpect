package httpexpect

import (
	"errors"
)

// Boolean provides methods to inspect attached bool value
// (Go representation of JSON boolean).
type Boolean struct {
	chain *chain
	value bool
}

// NewBoolean returns a new Boolean instance.
//
// reporter should not be nil.
//
// Example:
//
//	boolean := NewBoolean(t, true)
func NewBoolean(reporter Reporter, value bool) *Boolean {
	return newBoolean(newChainWithDefaults("Boolean()", reporter), value)
}

func newBoolean(parent *chain, val bool) *Boolean {
	return &Boolean{parent.clone(), val}
}

// Raw returns underlying value attached to Boolean.
// This is the value originally passed to NewBoolean.
//
// Example:
//
//	boolean := NewBoolean(t, true)
//	assert.Equal(t, true, boolean.Raw())
func (b *Boolean) Raw() bool {
	return b.value
}

// Path is similar to Value.Path.
func (b *Boolean) Path(path string) *Value {
	b.chain.enter("Path(%q)", path)
	defer b.chain.leave()

	return jsonPath(b.chain, b.value, path)
}

// Schema is similar to Value.Schema.
func (b *Boolean) Schema(schema interface{}) *Boolean {
	b.chain.enter("Schema()")
	defer b.chain.leave()

	jsonSchema(b.chain, b.value, schema)
	return b
}

// Equal succeeds if boolean is equal to given value.
//
// Example:
//
//	boolean := NewBoolean(t, true)
//	boolean.Equal(true)
func (b *Boolean) Equal(value bool) *Boolean {
	b.chain.enter("Equal()")
	defer b.chain.leave()

	if b.chain.failed() {
		return b
	}

	if !(b.value == value) {
		b.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{b.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: booleans are equal"),
			},
		})
	}

	return b
}

// NotEqual succeeds if boolean is not equal to given value.
//
// Example:
//
//	boolean := NewBoolean(t, true)
//	boolean.NotEqual(false)
func (b *Boolean) NotEqual(value bool) *Boolean {
	b.chain.enter("NotEqual()")
	defer b.chain.leave()

	if b.chain.failed() {
		return b
	}

	if !(b.value != value) {
		b.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{b.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: booleans are non-equal"),
			},
		})
	}

	return b
}

// True succeeds if boolean is true.
//
// Example:
//
//	boolean := NewBoolean(t, true)
//	boolean.True()
func (b *Boolean) True() *Boolean {
	b.chain.enter("True()")
	defer b.chain.leave()

	if b.chain.failed() {
		return b
	}

	if !(b.value == true) {
		b.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{b.value},
			Expected: &AssertionValue{true},
			Errors: []error{
				errors.New("expected: boolean is true"),
			},
		})
	}

	return b
}

// False succeeds if boolean is false.
//
// Example:
//
//	boolean := NewBoolean(t, false)
//	boolean.False()
func (b *Boolean) False() *Boolean {
	b.chain.enter("False()")
	defer b.chain.leave()

	if b.chain.failed() {
		return b
	}

	if !(b.value == false) {
		b.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{b.value},
			Expected: &AssertionValue{false},
			Errors: []error{
				errors.New("expected: boolean is false"),
			},
		})
	}

	return b
}
