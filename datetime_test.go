package httpexpect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateTimeFailed(t *testing.T) {
	chain := newMockChain(t)
	chain.fail(mockFailure())

	tm := time.Unix(0, 0)

	value := newDateTime(chain, tm)

	value.chain.assertFailed(t)

	value.Equal(tm)
	value.NotEqual(tm)
	value.Gt(tm)
	value.Ge(tm)
	value.Lt(tm)
	value.Le(tm)
	value.InRange(tm, tm)
	value.NotInRange(tm, tm)
	value.Zone()
	value.Year()
	value.Month()
	value.Day()
	value.WeekDay()
	value.YearDay()
	value.Hour()
	value.Minute()
	value.Second()
	value.Nanosecond()
	value.AsUTC()
	value.AsLocal()
}

func TestDatetimeConstructors(t *testing.T) {
	time := time.Unix(0, 1234)

	t.Run("Constructor without config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewDateTime(reporter, time)
		value.Equal(time)
		value.chain.assertNotFailed(t)
	})

	t.Run("Constructor with config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewDateTimeC(Config{
			Reporter: reporter,
		}, time)
		value.Equal(time)
		value.chain.assertNotFailed(t)
	})
}

func TestDateTimeEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDateTime(reporter, time.Unix(0, 1234))

	assert.True(t, time.Unix(0, 1234).Equal(value.Raw()))

	value.Equal(time.Unix(0, 1234))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Equal(time.Unix(0, 4321))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(time.Unix(0, 4321))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(time.Unix(0, 1234))
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDateTimeGreater(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDateTime(reporter, time.Unix(0, 1234))

	value.Gt(time.Unix(0, 1234-1))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Gt(time.Unix(0, 1234))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Ge(time.Unix(0, 1234-1))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Ge(time.Unix(0, 1234))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Ge(time.Unix(0, 1234+1))
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDateTimeLesser(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDateTime(reporter, time.Unix(0, 1234))

	value.Lt(time.Unix(0, 1234+1))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Lt(time.Unix(0, 1234))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Le(time.Unix(0, 1234+1))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Le(time.Unix(0, 1234))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Le(time.Unix(0, 1234-1))
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDateTimeInRange(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDateTime(reporter, time.Unix(0, 1234))

	value.InRange(time.Unix(0, 1234), time.Unix(0, 1234))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Unix(0, 1234), time.Unix(0, 1234))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Unix(0, 1234-1), time.Unix(0, 1234))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Unix(0, 1234-1), time.Unix(0, 1234))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Unix(0, 1234), time.Unix(0, 1234+1))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Unix(0, 1234), time.Unix(0, 1234+1))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Unix(0, 1234+1), time.Unix(0, 1234+2))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Unix(0, 1234+1), time.Unix(0, 1234+2))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Unix(0, 1234-2), time.Unix(0, 1234-1))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Unix(0, 1234-2), time.Unix(0, 1234-1))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Unix(0, 1234+1), time.Unix(0, 1234-1))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Unix(0, 1234+1), time.Unix(0, 1234-1))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestDateTimeGetters(t *testing.T) {
	reporter := newMockReporter(t)

	parsedTime, _ := time.Parse("Jan 2, 2006 at 3:04pm (MST)", "Dec 30, 0000 at 3:04pm (IST)")
	value := NewDateTime(reporter, parsedTime)

	value.chain.assertNotFailed(t)

	value.Zone().chain.assertNotFailed(t)
	value.Year().chain.assertNotFailed(t)
	value.Month().chain.assertNotFailed(t)
	value.Day().chain.assertNotFailed(t)
	value.WeekDay().chain.assertNotFailed(t)
	value.YearDay().chain.assertNotFailed(t)
	value.Hour().chain.assertNotFailed(t)
	value.Minute().chain.assertNotFailed(t)
	value.Second().chain.assertNotFailed(t)
	value.Nanosecond().chain.assertNotFailed(t)
	value.AsUTC().chain.assertNotFailed(t)
	value.AsLocal().chain.assertNotFailed(t)

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
