package httpexpect

import (
	"errors"
)

// Boolean provides methods to inspect attached bool value
// (Go representation of JSON boolean).
type Boolean struct {
	noCopy noCopy
	chain  *chain
	value  bool
}

// NewBoolean returns a new Boolean instance.
//
// If reporter is nil, the function panics.
//
// Example:
//
//	boolean := NewBoolean(t, true)
func NewBoolean(reporter Reporter, value bool) *Boolean {
	return newBoolean(newChainWithDefaults("Boolean()", reporter), value)
}

// NewBooleanC returns a new Boolean instance with config.
//
// Requirements for config are same as for WithConfig function.
//
// Example:
//
//	boolean := NewBooleanC(config, true)
func NewBooleanC(config Config, value bool) *Boolean {
	return newBoolean(newChainWithConfig("Boolean()", config.withDefaults()), value)
}

func newBoolean(parent *chain, val bool) *Boolean {
	return &Boolean{chain: parent.clone(), value: val}
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

// Decode unmarshals the underlying value attached to the Boolean to a target variable.
// target should be one of these:
//
// - pointer to an empty interface
// - pointer to a boolean
//
// Example:
//
//	value := NewBoolean(t, true)
//
//	var target bool
//	value.Decode(&target)
//
//	assert.Equal(t, true, target)
func (b *Boolean) Decode(target interface{}) *Boolean {
	opChain := b.chain.enter("Decode()")
	defer opChain.leave()

	if opChain.failed() {
		return b
	}

	canonDecode(opChain, b.value, target)
	return b
}

// Alias is similar to Value.Alias.
func (b *Boolean) Alias(name string) *Boolean {
	opChain := b.chain.enter("Alias(%q)", name)
	defer opChain.leave()

	b.chain.setAlias(name)
	return b
}

// Path is similar to Value.Path.
func (b *Boolean) Path(path string) *Value {
	opChain := b.chain.enter("Path(%q)", path)
	defer opChain.leave()

	return jsonPath(opChain, b.value, path)
}

// Schema is similar to Value.Schema.
func (b *Boolean) Schema(schema interface{}) *Boolean {
	opChain := b.chain.enter("Schema()")
	defer opChain.leave()

	jsonSchema(opChain, b.value, schema)
	return b
}

// Equal succeeds if boolean is equal to given value.
//
// Example:
//
//	boolean := NewBoolean(t, true)
//	boolean.Equal(true)
func (b *Boolean) Equal(value bool) *Boolean {
	opChain := b.chain.enter("Equal()")
	defer opChain.leave()

	if opChain.failed() {
		return b
	}

	if !(b.value == value) {
		opChain.fail(AssertionFailure{
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
	opChain := b.chain.enter("NotEqual()")
	defer opChain.leave()

	if opChain.failed() {
		return b
	}

	if b.value == value {
		opChain.fail(AssertionFailure{
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
	opChain := b.chain.enter("True()")
	defer opChain.leave()

	if opChain.failed() {
		return b
	}

	if !(b.value == true) {
		opChain.fail(AssertionFailure{
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
	opChain := b.chain.enter("False()")
	defer opChain.leave()

	if opChain.failed() {
		return b
	}

	if !(b.value == false) {
		opChain.fail(AssertionFailure{
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
