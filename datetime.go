package httpexpect

import (
	"errors"
	"time"
)

// DateTime provides methods to inspect attached time.Time value.
type DateTime struct {
	noCopy noCopy
	chain  *chain
	value  time.Time
}

// NewDateTime returns a new DateTime instance.
//
// If reporter is nil, the function panics.
//
// Example:
//
//	dt := NewDateTime(t, time.Now())
//	dt.Le(time.Now())
//
//	time.Sleep(time.Second)
//	dt.Lt(time.Now())
func NewDateTime(reporter Reporter, value time.Time) *DateTime {
	return newDateTime(newChainWithDefaults("DateTime()", reporter), value)
}

// NewDateTimeC returns a new DateTime instance with config.
//
// Requirements for config are same as for WithConfig function.
//
// Example:
//
//	dt := NewDateTimeC(config, time.Now())
//	dt.Le(time.Now())
//
//	time.Sleep(time.Second)
//	dt.Lt(time.Now())
func NewDateTimeC(config Config, value time.Time) *DateTime {
	return newDateTime(newChainWithConfig("DateTime()", config.withDefaults()), value)
}

func newDateTime(parent *chain, val time.Time) *DateTime {
	return &DateTime{chain: parent.clone(), value: val}
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

// GetZone returns a new String instance with datetime zone.
//
// Example:
//
//	tm, _ := time.Parse(time.UnixDate, "Fri Dec 30 15:04:05 IST 2022")
//	dt := NewDateTime(t, tm)
//	dt.GetZone().Equal("IST")
func (dt *DateTime) GetZone() *String {
	dt.chain.enter("GetZone()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return newString(dt.chain, "")
	}

	zone, _ := dt.value.Zone()
	return newString(dt.chain, zone)
}

// GetYear returns the year in which datetime occurs,
// in the range [0, 9999]
//
// Example:
//
//	tm, _ := time.Parse(time.UnixDate, "Fri Dec 30 15:04:05 IST 2022")
//	dt := NewDateTime(t, tm)
//	dt.GetYear().Equal(2022)
func (dt *DateTime) GetYear() *Number {
	dt.chain.enter("GetYear()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return newNumber(dt.chain, float64(0))
	}

	return newNumber(dt.chain, float64(dt.value.Year()))
}

// GetMonth returns the month of the year specified by datetime,
// in the range [1,12].
//
// Example:
//
//	tm, _ := time.Parse(time.UnixDate, "Fri Dec 30 15:04:05 IST 2022")
//	dt := NewDateTime(t, tm)
//	dt.GetMonth().Equal(12)
func (dt *DateTime) GetMonth() *Number {
	dt.chain.enter("GetMonth()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return newNumber(dt.chain, float64(0))
	}

	return newNumber(dt.chain, float64(dt.value.Month()))
}

// GetDay returns the day of the month specified datetime,
// in the range [1,31].
//
// Example:
//
//	tm, _ := time.Parse(time.UnixDate, "Fri Dec 30 15:04:05 IST 2022")
//	dt := NewDateTime(t, tm)
//	dt.GetDay().Equal(30)
func (dt *DateTime) GetDay() *Number {
	dt.chain.enter("GetDay()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return newNumber(dt.chain, float64(0))
	}

	return newNumber(dt.chain, float64(dt.value.Day()))
}

// Weekday returns the day of the week specified by datetime,
// in the range [0, 6], 0 corresponds to Sunday
//
// Example:
//
//	tm, _ := time.Parse(time.UnixDate, "Fri Dec 30 15:04:05 IST 2022")
//	dt := NewDateTime(t, tm)
//	dt.GetWeekDay().Equal(time.Friday)
func (dt *DateTime) GetWeekDay() *Number {
	dt.chain.enter("GetWeekDay()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return newNumber(dt.chain, float64(0))
	}

	return newNumber(dt.chain, float64(dt.value.Weekday()))
}

// GetYearDay returns the day of the year specified by datetime,
// in the range [1,365] for non-leap years,
// and [1,366] in leap years.
//
// Example:
//
//	tm, _ := time.Parse(time.UnixDate, "Fri Dec 30 15:04:05 IST 2022")
//	dt := NewDateTime(t, tm)
//	dt.GetYearDay().Equal(364)
func (dt *DateTime) GetYearDay() *Number {
	dt.chain.enter("GetYearDay()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return newNumber(dt.chain, float64(0))
	}

	return newNumber(dt.chain, float64(dt.value.YearDay()))
}

// GetHour returns the hour within the day specified by datetime,
// in the range [0, 23].
//
// Example:
//
//	tm, _ := time.Parse(time.UnixDate, "Fri Dec 30 15:04:05 IST 2022")
//	dt := NewDateTime(t, tm)
//	dt.GetHour().Equal(15)
func (dt *DateTime) GetHour() *Number {
	dt.chain.enter("GetHour()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return newNumber(dt.chain, float64(0))
	}

	return newNumber(dt.chain, float64(dt.value.Hour()))
}

// GetMinute returns the minute offset within the hour specified by datetime,
// in the range [0, 59].
//
// Example:
//
//	tm, _ := time.Parse(time.UnixDate, "Fri Dec 30 15:04:05 IST 2022")
//	dt := NewDateTime(t, tm)
//	dt.GetMinute().Equal(4)
func (dt *DateTime) GetMinute() *Number {
	dt.chain.enter("GetMinute()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return newNumber(dt.chain, float64(0))
	}

	return newNumber(dt.chain, float64(dt.value.Minute()))
}

// GetSecond returns the second offset within the minute specified by datetime,
// in the range [0, 59].
//
// Example:
//
//	tm, _ := time.Parse(time.UnixDate, "Fri Dec 30 15:04:05 IST 2022")
//	dt := NewDateTime(t, tm)
//	dt.GetSecond().Equal(5)
func (dt *DateTime) GetSecond() *Number {
	dt.chain.enter("GetSecond()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return newNumber(dt.chain, float64(0))
	}

	return newNumber(dt.chain, float64(dt.value.Second()))
}

// GetNanosecond returns the nanosecond offset within the second specified by datetime,
// in the range [0, 999999999].
//
// Example:
//
//	tm, _ := time.Parse(time.UnixDate, "Fri Dec 30 15:04:05 IST 2022")
//	dt := NewDateTime(t, tm)
//	dt.GetNanosecond().Equal(0)
func (dt *DateTime) GetNanosecond() *Number {
	dt.chain.enter("GetNanosecond()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return newNumber(dt.chain, float64(0))
	}

	return newNumber(dt.chain, float64(dt.value.Nanosecond()))
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

// AsUTC returns a new DateTime instance in UTC timeZone.
//
// Example:
//
//	tm, _ := time.Parse(time.UnixDate, "Fri Dec 30 15:04:05 IST 2022")
//	dt := NewDateTime(t, tm)
//	dt.AsUTC().Zone().Equal("UTC")
func (dt *DateTime) AsUTC() *DateTime {
	dt.chain.enter("AsUTC()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return dt
	}

	return newDateTime(dt.chain, dt.value.UTC())
}

// AsLocal returns a new DateTime instance in Local timeZone.
//
// Example:
//
//	tm, _ := time.Parse(time.UnixDate, "Fri Dec 30 15:04:05 IST 2022")
//	dt := NewDateTime(t, tm)
//	dt.AsLocal().Zone().Equal("IST")
func (dt *DateTime) AsLocal() *DateTime {
	dt.chain.enter("AsLocal()")
	defer dt.chain.leave()

	if dt.chain.failed() {
		return dt
	}

	return newDateTime(dt.chain, dt.value.Local())
}
