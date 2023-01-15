package httpexpect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDateTime_Failed(t *testing.T) {
	chain := newMockChain(t)
	chain.setFailed()

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
	value.GetZone()
	value.GetYear()
	value.GetMonth()
	value.GetDay()
	value.GetWeekDay()
	value.GetYearDay()
	value.GetHour()
	value.GetMinute()
	value.GetSecond()
	value.GetNanosecond()
	value.AsUTC()
	value.AsLocal()
	value.Alias("foo")
}

func TestDateTime_Constructors(t *testing.T) {
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

	t.Run("chain Constructor", func(t *testing.T) {
		chain := newMockChain(t)
		value := newDateTime(chain, time)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestDateTime_Alias(t *testing.T) {
	reporter := newMockReporter(t)
	value1 := NewDateTime(reporter, time.Unix(0, 1234))
	assert.Equal(t, []string{"DateTime()"}, value1.chain.context.Path)
	assert.Equal(t, []string{"DateTime()"}, value1.chain.context.AliasedPath)

	value2 := value1.Alias("foo")
	assert.Equal(t, []string{"DateTime()"}, value2.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value2.chain.context.AliasedPath)
}

func TestDateTime_Equal(t *testing.T) {
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

func TestDateTime_Greater(t *testing.T) {
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

func TestDateTime_Lesser(t *testing.T) {
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

func TestDateTime_InRange(t *testing.T) {
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

	parsedTime, _ := time.Parse(time.UnixDate, "FRI Dec 30 15:04:05 IST 2022")

	value := NewDateTime(reporter, parsedTime)

	value.chain.assertNotFailed(t)

	value.GetZone().chain.assertNotFailed(t)
	value.GetYear().chain.assertNotFailed(t)
	value.GetMonth().chain.assertNotFailed(t)
	value.GetDay().chain.assertNotFailed(t)
	value.GetWeekDay().chain.assertNotFailed(t)
	value.GetYearDay().chain.assertNotFailed(t)
	value.GetHour().chain.assertNotFailed(t)
	value.GetMinute().chain.assertNotFailed(t)
	value.GetSecond().chain.assertNotFailed(t)
	value.GetNanosecond().chain.assertNotFailed(t)
	value.AsUTC().chain.assertNotFailed(t)
	value.AsLocal().chain.assertNotFailed(t)

	expectedTime := parsedTime
	expectedZone, _ := expectedTime.Zone()
	assert.Equal(t, expectedZone, value.GetZone().Raw())
	assert.Equal(t, float64(expectedTime.Year()), value.GetYear().Raw())
	assert.Equal(t, float64(expectedTime.Month()), value.GetMonth().Raw())
	assert.Equal(t, float64(expectedTime.Day()), value.GetDay().Raw())
	assert.Equal(t, float64(expectedTime.Weekday()), value.GetWeekDay().Raw())
	assert.Equal(t, float64(expectedTime.YearDay()), value.GetYearDay().Raw())
	assert.Equal(t, float64(expectedTime.Hour()), value.GetHour().Raw())
	assert.Equal(t, float64(expectedTime.Minute()), value.GetMinute().Raw())
	assert.Equal(t, float64(expectedTime.Second()), value.GetSecond().Raw())
	assert.Equal(t, float64(expectedTime.Nanosecond()), value.GetNanosecond().Raw())
}
