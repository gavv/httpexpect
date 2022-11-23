package httpexpect

import (
	"errors"
	"math"
)

// Number provides methods to inspect attached float64 value
// (Go representation of JSON number).
type Number struct {
	chain *chain
	value float64
}

// NewNumber returns a new Number instance.
//
// reporter should not be nil.
//
// Example:
//
//	number := NewNumber(t, 123.4)
func NewNumber(reporter Reporter, value float64) *Number {
	return newNumber(newChainWithDefaults("Number()", reporter), value)
}

func newNumber(parent *chain, val float64) *Number {
	return &Number{parent.clone(), val}
}

// Raw returns underlying value attached to Number.
// This is the value originally passed to NewNumber.
//
// Example:
//
//	number := NewNumber(t, 123.4)
//	assert.Equal(t, 123.4, number.Raw())
func (n *Number) Raw() float64 {
	return n.value
}

// Path is similar to Value.Path.
func (n *Number) Path(path string) *Value {
	n.chain.enter("Path(%q)", path)
	defer n.chain.leave()

	return jsonPath(n.chain, n.value, path)
}

// Schema is similar to Value.Schema.
func (n *Number) Schema(schema interface{}) *Number {
	n.chain.enter("Schema()")
	defer n.chain.leave()

	jsonSchema(n.chain, n.value, schema)
	return n
}

// Equal succeeds if number is equal to given value.
//
// value should have numeric type convertible to float64. Before comparison,
// it is converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.Equal(float64(123))
//	number.Equal(int32(123))
func (n *Number) Equal(value interface{}) *Number {
	n.chain.enter("Equal()")
	defer n.chain.leave()

	if n.chain.failed() {
		return n
	}

	num, ok := canonNumber(n.chain, value)
	if !ok {
		return n
	}

	if !(n.value == num) {
		n.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: numbers are equal"),
			},
		})
	}

	return n
}

// NotEqual succeeds if number is not equal to given value.
//
// value should have numeric type convertible to float64. Before comparison,
// it is converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.NotEqual(float64(321))
//	number.NotEqual(int32(321))
func (n *Number) NotEqual(value interface{}) *Number {
	n.chain.enter("NotEqual()")
	defer n.chain.leave()

	if n.chain.failed() {
		return n
	}

	num, ok := canonNumber(n.chain, value)
	if !ok {
		return n
	}

	if !(n.value != num) {
		n.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: numbers are non-equal"),
			},
		})
	}

	return n
}

// EqualDelta succeeds if two numerals are within delta of each other.
//
// Example:
//
//	number := NewNumber(t, 123.0)
//	number.EqualDelta(123.2, 0.3)
func (n *Number) EqualDelta(value, delta float64) *Number {
	n.chain.enter("EqualDelta()")
	defer n.chain.leave()

	if n.chain.failed() {
		return n
	}

	if math.IsNaN(n.value) || math.IsNaN(value) || math.IsNaN(delta) {
		n.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{value},
			Delta:    &AssertionValue{delta},
			Errors: []error{
				errors.New("expected: numbers are comparable"),
			},
		})
		return n
	}

	diff := n.value - value

	if diff < -delta || diff > delta {
		n.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{value},
			Delta:    &AssertionValue{delta},
			Errors: []error{
				errors.New("expected: numbers lie within delta"),
			},
		})
		return n
	}

	return n
}

// NotEqualDelta succeeds if two numerals are not within delta of each other.
//
// Example:
//
//	number := NewNumber(t, 123.0)
//	number.NotEqualDelta(123.2, 0.1)
func (n *Number) NotEqualDelta(value, delta float64) *Number {
	n.chain.enter("NotEqualDelta()")
	defer n.chain.leave()

	if n.chain.failed() {
		return n
	}

	if math.IsNaN(n.value) || math.IsNaN(value) || math.IsNaN(delta) {
		n.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{value},
			Delta:    &AssertionValue{delta},
			Errors: []error{
				errors.New("expected: numbers are comparable"),
			},
		})
		return n
	}

	diff := n.value - value

	if !(diff < -delta || diff > delta) {
		n.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{value},
			Delta:    &AssertionValue{delta},
			Errors: []error{
				errors.New("expected: numbers do not lie within delta"),
			},
		})
		return n
	}

	return n
}

// Gt succeeds if number is greater than given value.
//
// value should have numeric type convertible to float64. Before comparison,
// it is converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.Gt(float64(122))
//	number.Gt(int32(122))
func (n *Number) Gt(value interface{}) *Number {
	n.chain.enter("Gt()")
	defer n.chain.leave()

	if n.chain.failed() {
		return n
	}

	num, ok := canonNumber(n.chain, value)
	if !ok {
		return n
	}

	if !(n.value > num) {
		n.chain.fail(AssertionFailure{
			Type:     AssertGt,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: number is larger than given value"),
			},
		})
	}

	return n
}

// Ge succeeds if number is greater than or equal to given value.
//
// value should have numeric type convertible to float64. Before comparison,
// it is converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.Ge(float64(122))
//	number.Ge(int32(122))
func (n *Number) Ge(value interface{}) *Number {
	n.chain.enter("Ge()")
	defer n.chain.leave()

	if n.chain.failed() {
		return n
	}

	num, ok := canonNumber(n.chain, value)
	if !ok {
		return n
	}

	if !(n.value >= num) {
		n.chain.fail(AssertionFailure{
			Type:     AssertGe,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: number is larger than or equal to given value"),
			},
		})
	}

	return n
}

// Lt succeeds if number is lesser than given value.
//
// value should have numeric type convertible to float64. Before comparison,
// it is converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.Lt(float64(124))
//	number.Lt(int32(124))
func (n *Number) Lt(value interface{}) *Number {
	n.chain.enter("Lt()")
	defer n.chain.leave()

	if n.chain.failed() {
		return n
	}

	num, ok := canonNumber(n.chain, value)
	if !ok {
		return n
	}

	if !(n.value < num) {
		n.chain.fail(AssertionFailure{
			Type:     AssertLt,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: number is less than given value"),
			},
		})
	}

	return n
}

// Le succeeds if number is lesser than or equal to given value.
//
// value should have numeric type convertible to float64. Before comparison,
// it is converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.Le(float64(124))
//	number.Le(int32(124))
func (n *Number) Le(value interface{}) *Number {
	n.chain.enter("Le()")
	defer n.chain.leave()

	if n.chain.failed() {
		return n
	}

	num, ok := canonNumber(n.chain, value)
	if !ok {
		return n
	}

	if !(n.value <= num) {
		n.chain.fail(AssertionFailure{
			Type:     AssertLe,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{num},
			Errors: []error{
				errors.New("expected: number is less than or equal to given value"),
			},
		})
	}

	return n
}

// InRange succeeds if number is within given range [min; max].
//
// min and max should have numeric type convertible to float64. Before comparison,
// they are converted to float64.
//
// Example:
//
//	number := NewNumber(t, 123)
//	number.InRange(float32(100), int32(200))  // success
//	number.InRange(100, 200)                  // success
//	number.InRange(123, 123)                  // success
func (n *Number) InRange(min, max interface{}) *Number {
	n.chain.enter("InRange()")
	defer n.chain.leave()

	if n.chain.failed() {
		return n
	}

	a, ok := canonNumber(n.chain, min)
	if !ok {
		return n
	}

	b, ok := canonNumber(n.chain, max)
	if !ok {
		return n
	}

	if !(n.value >= a && n.value <= b) {
		n.chain.fail(AssertionFailure{
			Type:     AssertInRange,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{AssertionRange{a, b}},
			Errors: []error{
				errors.New("expected: number is within given range"),
			},
		})
	}

	return n
}

// NotInRange succeeds if number is not within given range [min; max].
//
// min and max should have numeric type convertible to float64. Before comparison,
// they are converted to float64.
//
// Example:
//
//	number := NewNumber(t, 100)
//	number.NotInRange(0, 99)
//	number.NotInRange(101, 200)
func (n *Number) NotInRange(min, max interface{}) *Number {
	n.chain.enter("NotInRange()")
	defer n.chain.leave()

	if n.chain.failed() {
		return n
	}

	a, ok := canonNumber(n.chain, min)
	if !ok {
		return n
	}

	b, ok := canonNumber(n.chain, max)
	if !ok {
		return n
	}

	if n.value >= a && n.value <= b {
		n.chain.fail(AssertionFailure{
			Type:     AssertNotInRange,
			Actual:   &AssertionValue{n.value},
			Expected: &AssertionValue{AssertionRange{a, b}},
			Errors: []error{
				errors.New("expected: number is not within given range"),
			},
		})
	}

	return n
}
