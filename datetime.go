package httpexpect

import (
	"errors"
	"time"
)

// DateTime provides methods to inspect attached time.Time value.
type DateTime struct {
	chain *chain
	value time.Time
}

// NewDateTime returns a new DateTime instance.
//
// reporter should not be nil.
//
// Example:
//
//	dt := NewDateTime(reporter, time.Now())
//	dt.Le(time.Now())
//
//	time.Sleep(time.Second)
//	dt.Lt(time.Now())
func NewDateTime(reporter Reporter, value time.Time) *DateTime {
	return newDateTime(newChainWithDefaults("DateTime()", reporter), value)
}

func newDateTime(parent *chain, val time.Time) *DateTime {
	return &DateTime{parent.clone(), val}
}

// Raw returns underlying time.Time value attached to DateTime.
// This is the value originally passed to NewDateTime.
//
// Example:
//
//	dt := NewDateTime(t, timestamp)
//	assert.Equal(t, timestamp, dt.Raw())
func (dt *DateTime) Raw() time.Time {
	return dt.value
}

// Equal succeeds if DateTime is equal to given value.
//
// Example:
//
//	dt := NewDateTime(t, time.Unix(0, 1))
//	dt.Equal(time.Unix(0, 1))
func (dt *DateTime) Equal(value time.Time) *DateTime {
	dt.chain.enter("Equal()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return dt
	}

	if !dt.value.Equal(value) {
		dt.chain.fail(AssertionFailure{
			Type:     AssertEqual,
			Actual:   &AssertionValue{dt.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: time points are equal"),
			},
		})
	}

	return dt
}

// NotEqual succeeds if DateTime is not equal to given value.
//
// Example:
//
//	dt := NewDateTime(t, time.Unix(0, 1))
//	dt.NotEqual(time.Unix(0, 2))
func (dt *DateTime) NotEqual(value time.Time) *DateTime {
	dt.chain.enter("NotEqual()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return dt
	}

	if dt.value.Equal(value) {
		dt.chain.fail(AssertionFailure{
			Type:     AssertNotEqual,
			Actual:   &AssertionValue{dt.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: time points are non-equal"),
			},
		})
	}

	return dt
}

// Gt succeeds if DateTime is greater than given value.
//
// Example:
//
//	dt := NewDateTime(t, time.Unix(0, 2))
//	dt.Gt(time.Unix(0, 1))
func (dt *DateTime) Gt(value time.Time) *DateTime {
	dt.chain.enter("Gt()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return dt
	}

	if !dt.value.After(value) {
		dt.chain.fail(AssertionFailure{
			Type:     AssertGt,
			Actual:   &AssertionValue{dt.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: time point is after given time"),
			},
		})
	}

	return dt
}

// Ge succeeds if DateTime is greater than or equal to given value.
//
// Example:
//
//	dt := NewDateTime(t, time.Unix(0, 2))
//	dt.Ge(time.Unix(0, 1))
func (dt *DateTime) Ge(value time.Time) *DateTime {
	dt.chain.enter("Ge()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return dt
	}

	if !(dt.value.After(value) || dt.value.Equal(value)) {
		dt.chain.fail(AssertionFailure{
			Type:     AssertGe,
			Actual:   &AssertionValue{dt.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: time point is after or equal to given time"),
			},
		})
	}

	return dt
}

// Lt succeeds if DateTime is lesser than given value.
//
// Example:
//
//	dt := NewDateTime(t, time.Unix(0, 1))
//	dt.Lt(time.Unix(0, 2))
func (dt *DateTime) Lt(value time.Time) *DateTime {
	dt.chain.enter("Lt()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return dt
	}

	if !dt.value.Before(value) {
		dt.chain.fail(AssertionFailure{
			Type:     AssertLt,
			Actual:   &AssertionValue{dt.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: time point is before given time"),
			},
		})
	}

	return dt
}

// Le succeeds if DateTime is lesser than or equal to given value.
//
// Example:
//
//	dt := NewDateTime(t, time.Unix(0, 1))
//	dt.Le(time.Unix(0, 2))
func (dt *DateTime) Le(value time.Time) *DateTime {
	dt.chain.enter("Le()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return dt
	}

	if !(dt.value.Before(value) || dt.value.Equal(value)) {
		dt.chain.fail(AssertionFailure{
			Type:     AssertLe,
			Actual:   &AssertionValue{dt.value},
			Expected: &AssertionValue{value},
			Errors: []error{
				errors.New("expected: time point is before or equal to given time"),
			},
		})
	}

	return dt
}

// InRange succeeds if DateTime is within given range [min; max].
//
// Example:
//
//	dt := NewDateTime(t, time.Unix(0, 2))
//	dt.InRange(time.Unix(0, 1), time.Unix(0, 3))
//	dt.InRange(time.Unix(0, 2), time.Unix(0, 2))
func (dt *DateTime) InRange(min, max time.Time) *DateTime {
	dt.chain.enter("InRange()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return dt
	}

	if !((dt.value.After(min) || dt.value.Equal(min)) &&
		(dt.value.Before(max) || dt.value.Equal(max))) {
		dt.chain.fail(AssertionFailure{
			Type:     AssertInRange,
			Actual:   &AssertionValue{dt.value},
			Expected: &AssertionValue{AssertionRange{min, max}},
			Errors: []error{
				errors.New("expected: time point is within given range"),
			},
		})
	}

	return dt
}

// NotInRange succeeds if DateTime is not within given range [min; max].
//
// Example:
//
//	dt := NewDateTime(t, time.Unix(0, 10))
//	dt.NotInRange(time.Unix(0, 1), time.Unix(0, 9))
//	dt.NotInRange(time.Unix(0, 11), time.Unix(0, 20))
func (dt *DateTime) NotInRange(min, max time.Time) *DateTime {
	dt.chain.enter("NotInRange()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return dt
	}

	if (dt.value.After(min) || dt.value.Equal(min)) &&
		(dt.value.Before(max) || dt.value.Equal(max)) {
		dt.chain.fail(AssertionFailure{
			Type:     AssertNotInRange,
			Actual:   &AssertionValue{dt.value},
			Expected: &AssertionValue{AssertionRange{min, max}},
			Errors: []error{
				errors.New("expected: time point is not within given range"),
			},
		})
	}

	return dt
}
