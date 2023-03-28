package httpexpect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateTime_FailedChain(t *testing.T) {
	chain := newFailedChain(t)

	tm := time.Unix(0, 0)
	value := newDateTime(chain, tm)
	value.chain.assert(t, failure)

	value.Alias("foo")

	value.IsEqual(tm)
	value.NotEqual(tm)
	value.InRange(tm, tm)
	value.NotInRange(tm, tm)
	value.InList(tm, tm)
	value.NotInList(tm, tm)
	value.Gt(tm)
	value.Ge(tm)
	value.Lt(tm)
	value.Le(tm)

	value.Zone().chain.assert(t, failure)
	value.Year().chain.assert(t, failure)
	value.Month().chain.assert(t, failure)
	value.Day().chain.assert(t, failure)
	value.WeekDay().chain.assert(t, failure)
	value.YearDay().chain.assert(t, failure)
	value.Hour().chain.assert(t, failure)
	value.Minute().chain.assert(t, failure)
	value.Second().chain.assert(t, failure)
	value.Nanosecond().chain.assert(t, failure)

	value.AsUTC().chain.assert(t, failure)
	value.AsLocal().chain.assert(t, failure)
}

func TestDateTime_Constructors(t *testing.T) {
	time := time.Unix(0, 1234)

	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewDateTime(reporter, time)
		value.IsEqual(time)
		value.chain.assert(t, success)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewDateTimeC(Config{
			Reporter: reporter,
		}, time)
		value.IsEqual(time)
		value.chain.assert(t, success)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newDateTime(chain, time)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestDateTime_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDateTime(reporter, time.Unix(0, 1234))
	assert.Equal(t, []string{"DateTime()"}, value.chain.context.Path)
	assert.Equal(t, []string{"DateTime()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"DateTime()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)
}

func TestDateTime_Getters(t *testing.T) {
	reporter := newMockReporter(t)

	parsedTime, _ := time.Parse(time.UnixDate, "FRI Dec 30 15:04:05 IST 2022")

	value := NewDateTime(reporter, parsedTime)

	value.chain.assert(t, success)

	value.Zone().chain.assert(t, success)
	value.Year().chain.assert(t, success)
	value.Month().chain.assert(t, success)
	value.Day().chain.assert(t, success)
	value.WeekDay().chain.assert(t, success)
	value.YearDay().chain.assert(t, success)
	value.Hour().chain.assert(t, success)
	value.Minute().chain.assert(t, success)
	value.Second().chain.assert(t, success)
	value.Nanosecond().chain.assert(t, success)
	value.AsUTC().chain.assert(t, success)
	value.AsLocal().chain.assert(t, success)

	expectedTime := parsedTime
	expectedZone, _ := expectedTime.Zone()

	assert.Equal(t, expectedZone, value.Zone().Raw())
	assert.Equal(t, float64(expectedTime.Year()), value.Year().Raw())
	assert.Equal(t, float64(expectedTime.Month()), value.Month().Raw())
	assert.Equal(t, float64(expectedTime.Day()), value.Day().Raw())
	assert.Equal(t, float64(expectedTime.Weekday()), value.WeekDay().Raw())
	assert.Equal(t, float64(expectedTime.YearDay()), value.YearDay().Raw())
	assert.Equal(t, float64(expectedTime.Hour()), value.Hour().Raw())
	assert.Equal(t, float64(expectedTime.Minute()), value.Minute().Raw())
	assert.Equal(t, float64(expectedTime.Second()), value.Second().Raw())
	assert.Equal(t, float64(expectedTime.Nanosecond()), value.Nanosecond().Raw())
}

func TestDateTime_IsEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDateTime(reporter, time.Unix(0, 1234))

	assert.True(t, time.Unix(0, 1234).Equal(value.Raw()))

	value.IsEqual(time.Unix(0, 1234))
	value.chain.assert(t, success)
	value.chain.clear()

	value.IsEqual(time.Unix(0, 4321))
	value.chain.assert(t, failure)
	value.chain.clear()

	value.NotEqual(time.Unix(0, 4321))
	value.chain.assert(t, success)
	value.chain.clear()

	value.NotEqual(time.Unix(0, 1234))
	value.chain.assert(t, failure)
	value.chain.clear()
}

func TestDateTime_IsGreater(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDateTime(reporter, time.Unix(0, 1234))

	value.Gt(time.Unix(0, 1234-1))
	value.chain.assert(t, success)
	value.chain.clear()

	value.Gt(time.Unix(0, 1234))
	value.chain.assert(t, failure)
	value.chain.clear()

	value.Ge(time.Unix(0, 1234-1))
	value.chain.assert(t, success)
	value.chain.clear()

	value.Ge(time.Unix(0, 1234))
	value.chain.assert(t, success)
	value.chain.clear()

	value.Ge(time.Unix(0, 1234+1))
	value.chain.assert(t, failure)
	value.chain.clear()
}

func TestDateTime_IsLesser(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDateTime(reporter, time.Unix(0, 1234))

	value.Lt(time.Unix(0, 1234+1))
	value.chain.assert(t, success)
	value.chain.clear()

	value.Lt(time.Unix(0, 1234))
	value.chain.assert(t, failure)
	value.chain.clear()

	value.Le(time.Unix(0, 1234+1))
	value.chain.assert(t, success)
	value.chain.clear()

	value.Le(time.Unix(0, 1234))
	value.chain.assert(t, success)
	value.chain.clear()

	value.Le(time.Unix(0, 1234-1))
	value.chain.assert(t, failure)
	value.chain.clear()
}

func TestDateTime_InRange(t *testing.T) {
	cases := []struct {
		name           string
		value          time.Time
		min            time.Time
		max            time.Time
		wantInRange    chainResult
		wantNotInRange chainResult
	}{
		{
			name:           "value equal to both min and max",
			value:          time.Unix(0, 1234),
			min:            time.Unix(0, 1234),
			max:            time.Unix(0, 1234),
			wantInRange:    success,
			wantNotInRange: failure,
		},
		{
			name:           "value after min and equal to max",
			value:          time.Unix(0, 1234),
			min:            time.Unix(0, 1234-1),
			max:            time.Unix(0, 1234),
			wantInRange:    success,
			wantNotInRange: failure,
		},
		{
			name:           "value equal to min and before max",
			value:          time.Unix(0, 1234),
			min:            time.Unix(0, 1234),
			max:            time.Unix(0, 1234+1),
			wantInRange:    success,
			wantNotInRange: failure,
		},
		{
			name:           "value before range",
			value:          time.Unix(0, 1234),
			min:            time.Unix(0, 1234+1),
			max:            time.Unix(0, 1234+2),
			wantInRange:    failure,
			wantNotInRange: success,
		},
		{
			name:           "value after range",
			value:          time.Unix(0, 1234),
			min:            time.Unix(0, 1234-2),
			max:            time.Unix(0, 1234-1),
			wantInRange:    failure,
			wantNotInRange: success,
		},
		{
			name:           "invalid range",
			value:          time.Unix(0, 1234),
			min:            time.Unix(0, 1234+1),
			max:            time.Unix(0, 1234-1),
			wantInRange:    failure,
			wantNotInRange: success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewDateTime(reporter, tc.value).InRange(tc.min, tc.max).
				chain.assert(t, tc.wantInRange)

			NewDateTime(reporter, tc.value).NotInRange(tc.min, tc.max).
				chain.assert(t, tc.wantNotInRange)
		})
	}
}

func TestDateTime_InList(t *testing.T) {
	cases := []struct {
		name          string
		value         time.Time
		list          []time.Time
		wantInList    chainResult
		wantNotInList chainResult
	}{
		{
			name:          "empty list",
			value:         time.Unix(0, 1234),
			list:          []time.Time{},
			wantInList:    failure,
			wantNotInList: failure,
		},
		{
			name:          "value present in list",
			value:         time.Unix(0, 1234),
			list:          []time.Time{time.Unix(0, 1234), time.Unix(0, 1234+1)},
			wantInList:    success,
			wantNotInList: failure,
		},
		{
			name:          "value not present in list",
			value:         time.Unix(0, 1234),
			list:          []time.Time{time.Unix(0, 1234-1), time.Unix(0, 1234+1)},
			wantInList:    failure,
			wantNotInList: success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewDateTime(reporter, tc.value).InList(tc.list...).
				chain.assert(t, tc.wantInList)

			NewDateTime(reporter, tc.value).NotInList(tc.list...).
				chain.assert(t, tc.wantNotInList)
		})
	}
}
