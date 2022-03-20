package httpexpect

import (
	"time"
)

// Duration provides methods to inspect attached time.Duration value.
type Duration struct {
	chain chain
	value *time.Duration
}

// NewDuration returns a new Duration object given a reporter used to report
// failures and time.Duration value to be inspected.
//
// reporter should not be nil.
//
// Example:
//   d := NewDuration(reporter, time.Second)
//   d.Le(time.Minute)
func NewDuration(reporter Reporter, value time.Duration) *Duration {
	return &Duration{makeChain(reporter), &value}
}

// Raw returns underlying time.Duration value attached to Duration.
// This is the value originally passed to NewDuration.
//
// Example:
//  d := NewDuration(t, duration)
//  assert.Equal(t, timestamp, d.Raw())
func (d *Duration) Raw() time.Duration {
	if d.value == nil {
		return 0
	}
	return *d.value
}

// IsSet succeeds if Duration is set.
//
// Example:
//  d := NewDuration(t, time.Second)
//  d.IsSet()
func (d *Duration) IsSet() *Duration {
	if d.value == nil {
		d.chain.fail(Failure{
			AssertionName: "Duration.IsSet",
			AssertType:    FailureAssertNotNil,
		})
	}
	return d
}

// NotSet succeeds if Duration is not set.
func (d *Duration) NotSet() *Duration {
	if d.value != nil {
		d.chain.fail(Failure{
			AssertionName: "Duration.NotSet",
			AssertType:    FailureAssertNil,
		})
	}
	return d
}

// Equal succeeds if Duration is equal to given value.
//
// Example:
//  d := NewDuration(t, time.Second)
//  d.Equal(time.Second)
func (d *Duration) Equal(value time.Duration) *Duration {
	if d.IsSet().chain.failed() {
		return d
	}

	if !(*d.value == value) {
		d.chain.fail(Failure{
			AssertionName: "Duration.Equal",
			AssertType:    FailureAssertEqual,
			Expected:      value,
			Actual:        *d.value,
		})
	}
	return d
}

// NotEqual succeeds if Duration is not equal to given value.
//
// Example:
//  d := NewDuration(t, time.Second)
//  d.NotEqual(time.Minute)
func (d *Duration) NotEqual(value time.Duration) *Duration {
	if d.IsSet().chain.failed() {
		return d
	}

	if !(*d.value != value) {
		d.chain.fail(Failure{
			AssertionName: "Duration.NotEqual",
			Expected:      value,
			Actual:        *d.value,
			AssertType:    FailureAssertNotEqual,
		})
	}
	return d
}

// Gt succeeds if Duration is greater than given value.
//
// Example:
//  d := NewDuration(t, time.Minute)
//  d.Gt(time.Second)
func (d *Duration) Gt(value time.Duration) *Duration {
	if d.IsSet().chain.failed() {
		return d
	}

	if !(*d.value > value) {
		d.chain.fail(Failure{
			AssertionName: "Duration.Gt",
			AssertType:    FailureAssertGt,
			Expected:      value,
			Actual:        *d.value,
		})
	}
	return d
}

// Ge succeeds if Duration is greater than or equal to given value.
//
// Example:
//  d := NewDuration(t, time.Minute)
//  d.Ge(time.Second)
func (d *Duration) Ge(value time.Duration) *Duration {
	if d.IsSet().chain.failed() {
		return d
	}

	if !(*d.value >= value) {
		d.chain.fail(Failure{
			AssertionName: "Duration.Ge",
			AssertType:    FailureAssertGe,
			Expected:      value,
			Actual:        *d.value,
		})
	}
	return d
}

// Lt succeeds if Duration is lesser than given value.
//
// Example:
//  d := NewDuration(t, time.Second)
//  d.Lt(time.Minute)
func (d *Duration) Lt(value time.Duration) *Duration {
	if d.IsSet().chain.failed() {
		return d
	}

	if !(*d.value < value) {
		d.chain.fail(Failure{
			AssertionName: "Duration.Lt",
			AssertType:    FailureAssertLt,
			Expected:      value,
			Actual:        *d.value,
		})
	}
	return d
}

// Le succeeds if Duration is lesser than or equal to given value.
//
// Example:
//  d := NewDuration(t, time.Second)
//  d.Le(time.Minute)
func (d *Duration) Le(value time.Duration) *Duration {
	if d.IsSet().chain.failed() {
		return d
	}

	if !(*d.value <= value) {
		d.chain.fail(Failure{
			AssertionName: "Duration.Le",
			AssertType:    FailureAssertLe,
			Expected:      value,
			Actual:        *d.value,
		})
	}
	return d
}

// InRange succeeds if Duration is in given range [min; max].
//
// Example:
//  d := NewDuration(t, time.Minute)
//  d.InRange(time.Second, time.Hour)
//  d.InRange(time.Minute, time.Minute)
func (d *Duration) InRange(min, max time.Duration) *Duration {
	if d.IsSet().chain.failed() {
		return d
	}

	if !(*d.value >= min && *d.value <= max) {
		d.chain.fail(Failure{
			AssertionName:   "Duration.InRange",
			AssertType:      FailureAssertInRange,
			ExpectedInRange: []interface{}{min, max},
			Actual:          *d.value,
		})
	}
	return d
}
