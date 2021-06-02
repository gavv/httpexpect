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
		failure := Failure{
			assertionName: "duration",
			assertType:    failureAssertIsSet,
		}
		d.chain.fail(failure)
	}
	return d
}

// NotSet succeeds if Duration is not set.
func (d *Duration) NotSet() *Duration {
	if d.value != nil {
		failure := Failure{
			assertionName: "duration",
			assertType:    failureAssertNotSet,
		}
		d.chain.fail(failure)
	}
	return d
}

// Equal succeeds if Duration is equal to given value.
//
// Example:
//  d := NewDuration(t, time.Second)
//  d.Equal(time.Second)
func (d *Duration) Equal(value time.Duration) *Duration {
	d.IsSet()

	if !(*d.value == value) {
		failure := Failure{
			assertionName: "duration",
			assertType:    failureAssertEqual,
			expected:      value,
			actual:        *d.value,
		}
		d.chain.fail(failure)
	}
	return d
}

// NotEqual succeeds if Duration is not equal to given value.
//
// Example:
//  d := NewDuration(t, time.Second)
//  d.NotEqual(time.Minute)
func (d *Duration) NotEqual(value time.Duration) *Duration {
	d.IsSet()

	if !(*d.value != value) {
		failure := Failure{
			assertionName: "duration",
			expected:      value,
			actual:        *d.value,
			assertType:    failureAssertNotEqual,
		}
		d.chain.fail(failure)
	}
	return d
}

// Gt succeeds if Duration is greater than given value.
//
// Example:
//  d := NewDuration(t, time.Minute)
//  d.Gt(time.Second)
func (d *Duration) Gt(value time.Duration) *Duration {
	d.IsSet()

	if !(*d.value > value) {
		failure := Failure{
			assertionName: "duration",
			assertType:    failureAssertGt,
			expected:      value,
			actual:        *d.value,
		}
		d.chain.fail(failure)
	}
	return d
}

// Ge succeeds if Duration is greater than or equal to given value.
//
// Example:
//  d := NewDuration(t, time.Minute)
//  d.Ge(time.Second)
func (d *Duration) Ge(value time.Duration) *Duration {
	d.IsSet()

	if !(*d.value >= value) {
		failure := Failure{
			assertionName: "duration",
			assertType:    failureAssertGe,
			expected:      value,
			actual:        *d.value,
		}
		d.chain.fail(failure)
	}
	return d
}

// Lt succeeds if Duration is lesser than given value.
//
// Example:
//  d := NewDuration(t, time.Second)
//  d.Lt(time.Minute)
func (d *Duration) Lt(value time.Duration) *Duration {
	d.IsSet()

	if !(*d.value < value) {
		failure := Failure{
			assertionName: "duration",
			assertType:    failureAssertLt,
			expected:      value,
			actual:        *d.value,
		}
		d.chain.fail(failure)
	}
	return d
}

// Le succeeds if Duration is lesser than or equal to given value.
//
// Example:
//  d := NewDuration(t, time.Second)
//  d.Le(time.Minute)
func (d *Duration) Le(value time.Duration) *Duration {
	d.IsSet()

	if !(*d.value <= value) {
		failure := Failure{
			assertionName: "duration",
			assertType:    failureAssertLe,
			expected:      value,
			actual:        *d.value,
		}
		d.chain.fail(failure)
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
	d.IsSet()

	if !(*d.value >= min && *d.value <= max) {
		failure := Failure{
			assertionName:   "duration",
			assertType:      failureAssertInRange,
			expectedInRange: []interface{}{min, max},
			actual:          *d.value,
		}
		d.chain.fail(failure)
	}
	return d
}
