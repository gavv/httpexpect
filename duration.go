package httpexpect

import (
	"errors"
	"time"
)

// Duration provides methods to inspect attached time.Duration value.
type Duration struct {
	noCopy noCopy
	chain  *chain
	value  *time.Duration
}

// NewDuration returns a new Duration instance.
//
// If reporter is nil, the function panics.
//
// Example:
//
//	d := NewDuration(t, time.Second)
//	d.Le(time.Minute)
func NewDuration(reporter Reporter, value time.Duration) *Duration {
	return newDuration(newChainWithDefaults("Duration()", reporter), &value)
}

// NewDurationC returns a new Duration instance with config.
//
// Requirements for config are same as for WithConfig function.
//
// Example:
//
//	d := NewDurationC(config, time.Second)
//	d.Le(time.Minute)
func NewDurationC(config Config, value time.Duration) *Duration {
	return newDuration(newChainWithConfig("Duration()", config.withDefaults()), &value)
}

func newDuration(parent *chain, val *time.Duration) *Duration {
	return &Duration{chain: parent.clone(), value: val}
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

// Deprecated: support for unset durations will be removed. The only method that
// can create unset duration is Cookie.MaxAge. Instead of Cookie.MaxAge().IsSet(),
// please use Cookie.HaveMaxAge().
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

// Deprecated: support for unset durations will be removed. The only method that
// can create unset duration is Cookie.MaxAge. Instead of Cookie.MaxAge().NotSet(),
// please use Cookie.NotHaveMaxAge().
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

// InRange succeeds if Duration is within given range [min; max].
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

// NotInRange succeeds if Duration is not within given range [min; max].
//
// Example:
//
//	d := NewDuration(t, time.Minute*10)
//	d.NotInRange(time.Minute, time.Minute-time.Nanosecond)
//	d.NotInRange(time.Minute+time.Nanosecond, time.Minute*10)
func (d *Duration) NotInRange(min, max time.Duration) *Duration {
	d.chain.enter("NotInRange()")
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

	if *d.value >= min && *d.value <= max {
		d.chain.fail(AssertionFailure{
			Type:     AssertNotInRange,
			Actual:   &AssertionValue{d.value},
			Expected: &AssertionValue{AssertionRange{min, max}},
			Errors: []error{
				errors.New("expected: duration is not within given range"),
			},
		})
	}

	return d
}
