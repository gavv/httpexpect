package httpexpect

import (
	"errors"
	"time"
)

// Duration provides methods to inspect attached time.Duration value.
type Duration struct {
	chain *chain
	value *time.Duration
}

// NewDuration returns a new Duration object given a reporter used to report
// failures and time.Duration value to be inspected.
//
// reporter should not be nil.
//
// Example:
//
//	d := NewDuration(reporter, time.Second)
//	d.Le(time.Minute)
func NewDuration(reporter Reporter, value time.Duration) *Duration {
	return newDuration(newDefaultChain("Duration()", reporter), &value)
}

func newDuration(parent *chain, val *time.Duration) *Duration {
	return &Duration{parent.clone(), val}
}

// Raw returns underlying time.Duration value attached to Duration.
// This is the value originally passed to NewDuration.
//
// Example:
//
//	d := NewDuration(t, duration)
//	assert.Equal(t, timestamp, d.Raw())
func (d *Duration) Raw() time.Duration {
	if d.value == nil {
		return 0
	}
	return *d.value
}

// IsSet succeeds if Duration is set.
//
// Example:
//
//	d := NewDuration(t, time.Second)
//	d.IsSet()
func (d *Duration) IsSet() *Duration {
	d.chain.enter("IsSet()")
	defer d.chain.leave()

	if d.chain.failed() {
		return d
	}

	if d.value == nil {
		d.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{d.value},
			Errors: []error{
				errors.New("expected: duration is present"),
			},
		})
	}

	return d
}

// NotSet succeeds if Duration is not set.
func (d *Duration) NotSet() *Duration {
	d.chain.enter("NotSet()")
	defer d.chain.leave()

	if d.chain.failed() {
		return d
	}

	if !(d.value == nil) {
		d.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{d.value},
			Errors: []error{
				errors.New("expected: duration is not present"),
			},
		})
	}

	return d
}

// Equal succeeds if Duration is equal to given value.
//
// Example:
//
//	d := NewDuration(t, time.Second)
//	d.Equal(time.Second)
func (d *Duration) Equal(value time.Duration) *Duration {
	d.chain.enter("Equal()")
	defer d.chain.leave()

	if d.chain.failed() {
		return d
	}

	if d.value == nil {
		d.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{d.value},
			Errors: []error{
				errors.New("expected: duration is present"),
			},
		})
		return d
	}

	if !(*d.value == value) {
		d.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{d.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: durations are equal"),
			},
		})
	}

	return d
}

// NotEqual succeeds if Duration is not equal to given value.
//
// Example:
//
//	d := NewDuration(t, time.Second)
//	d.NotEqual(time.Minute)
func (d *Duration) NotEqual(value time.Duration) *Duration {
	d.chain.enter("NotEqual()")
	defer d.chain.leave()

	if d.chain.failed() {
		return d
	}

	if d.value == nil {
		d.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{d.value},
			Errors: []error{
				errors.New("expected: duration is present"),
			},
		})
		return d
	}

	if !(*d.value != value) {
		d.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{d.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: durations are non-equal"),
			},
		})
	}

	return d
}

// Gt succeeds if Duration is greater than given value.
//
// Example:
//
//	d := NewDuration(t, time.Minute)
//	d.Gt(time.Second)
func (d *Duration) Gt(value time.Duration) *Duration {
	d.chain.enter("Gt()")
	defer d.chain.leave()

	if d.chain.failed() {
		return d
	}

	if d.value == nil {
		d.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{d.value},
			Errors: []error{
				errors.New("expected: duration is present"),
			},
		})
		return d
	}

	if !(*d.value > value) {
		d.chain.fail(AssertionFailure{
			Type:     AssertGt,
			Actual:   &AssertionValue{d.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: duration is larger than given value"),
			},
		})
	}

	return d
}

// Ge succeeds if Duration is greater than or equal to given value.
//
// Example:
//
//	d := NewDuration(t, time.Minute)
//	d.Ge(time.Second)
func (d *Duration) Ge(value time.Duration) *Duration {
	d.chain.enter("Ge()")
	defer d.chain.leave()

	if d.chain.failed() {
		return d
	}

	if d.value == nil {
		d.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{d.value},
			Errors: []error{
				errors.New("expected: duration is present"),
			},
		})
		return d
	}

	if !(*d.value >= value) {
		d.chain.fail(AssertionFailure{
			Type:     AssertGe,
			Actual:   &AssertionValue{d.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: duration is larger than or equal to given value"),
			},
		})
	}

	return d
}

// Lt succeeds if Duration is lesser than given value.
//
// Example:
//
//	d := NewDuration(t, time.Second)
//	d.Lt(time.Minute)
func (d *Duration) Lt(value time.Duration) *Duration {
	d.chain.enter("Lt()")
	defer d.chain.leave()

	if d.chain.failed() {
		return d
	}

	if d.value == nil {
		d.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{d.value},
			Errors: []error{
				errors.New("expected: duration is present"),
			},
		})
		return d
	}

	if !(*d.value < value) {
		d.chain.fail(AssertionFailure{
			Type:     AssertLt,
			Actual:   &AssertionValue{d.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: duration is less than given value"),
			},
		})
	}

	return d
}

// Le succeeds if Duration is lesser than or equal to given value.
//
// Example:
//
//	d := NewDuration(t, time.Second)
//	d.Le(time.Minute)
func (d *Duration) Le(value time.Duration) *Duration {
	d.chain.enter("Le()")
	defer d.chain.leave()

	if d.chain.failed() {
		return d
	}

	if d.value == nil {
		d.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{d.value},
			Errors: []error{
				errors.New("expected: duration is present"),
			},
		})
		return d
	}

	if !(*d.value <= value) {
		d.chain.fail(AssertionFailure{
			Type:     AssertLe,
			Actual:   &AssertionValue{d.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: duration is less than or equal to given value"),
			},
		})
	}

	return d
}

// InRange succeeds if Duration is in given range [min; max].
//
// Example:
//
//	d := NewDuration(t, time.Minute)
//	d.InRange(time.Second, time.Hour)
//	d.InRange(time.Minute, time.Minute)
func (d *Duration) InRange(min, max time.Duration) *Duration {
	d.chain.enter("InRange()")
	defer d.chain.leave()

	if d.chain.failed() {
		return d
	}

	if d.value == nil {
		d.chain.fail(AssertionFailure{
			Type:   AssertNotNil,
			Actual: &AssertionValue{d.value},
			Errors: []error{
				errors.New("expected: duration is present"),
			},
		})
		return d
	}

	if !(*d.value >= min && *d.value <= max) {
		d.chain.fail(AssertionFailure{
			Type:     AssertInRange,
			Actual:   &AssertionValue{d.value},
			Expected: &AssertionValue{AssertionRange{min, max}},
			Errors: []error{
				errors.New("expected: duration is within given range"),
			},
		})
	}

	return d
}
