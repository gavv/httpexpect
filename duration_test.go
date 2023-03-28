package httpexpect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDuration_FailedChain(t *testing.T) {
	chain := newFailedChain(t)

	tm := time.Second
	value := newDuration(chain, &tm)
	value.chain.assert(t, failure)

	value.Alias("foo")
	value.IsEqual(tm)
	value.NotEqual(tm)
	value.InRange(tm, tm)
	value.NotInRange(tm, tm)
	value.InList(tm)
	value.NotInList(tm)
	value.Gt(tm)
	value.Ge(tm)
	value.Lt(tm)
	value.Le(tm)
}

func TestDuration_Constructors(t *testing.T) {
	tm := time.Second

	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewDuration(reporter, tm)
		value.IsEqual(tm)
		value.chain.assert(t, success)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewDurationC(Config{
			Reporter: reporter,
		}, tm)
		value.IsEqual(tm)
		value.chain.assert(t, success)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newDuration(chain, &tm)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestDuration_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)
	assert.Equal(t, []string{"Duration()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Duration()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Duration()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)
}

func TestDuration_IsEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	assert.Equal(t, time.Second, value.Raw())

	value.IsEqual(time.Second)
	value.chain.assert(t, success)
	value.chain.clear()

	value.IsEqual(time.Minute)
	value.chain.assert(t, failure)
	value.chain.clear()

	value.NotEqual(time.Minute)
	value.chain.assert(t, success)
	value.chain.clear()

	value.NotEqual(time.Second)
	value.chain.assert(t, failure)
	value.chain.clear()
}

func TestDuration_IsGreater(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	value.Gt(time.Second - 1)
	value.chain.assert(t, success)
	value.chain.clear()

	value.Gt(time.Second)
	value.chain.assert(t, failure)
	value.chain.clear()

	value.Ge(time.Second - 1)
	value.chain.assert(t, success)
	value.chain.clear()

	value.Ge(time.Second)
	value.chain.assert(t, success)
	value.chain.clear()

	value.Ge(time.Second + 1)
	value.chain.assert(t, failure)
	value.chain.clear()
}

func TestDuration_IsLesser(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	value.Lt(time.Second + 1)
	value.chain.assert(t, success)
	value.chain.clear()

	value.Lt(time.Second)
	value.chain.assert(t, failure)
	value.chain.clear()

	value.Le(time.Second + 1)
	value.chain.assert(t, success)
	value.chain.clear()

	value.Le(time.Second)
	value.chain.assert(t, success)
	value.chain.clear()

	value.Le(time.Second - 1)
	value.chain.assert(t, failure)
	value.chain.clear()
}

func TestDuration_InRange(t *testing.T) {
	cases := []struct {
		name           string
		value          time.Duration
		min            time.Duration
		max            time.Duration
		wantInRange    chainResult
		wantNotInRange chainResult
	}{
		{
			name:           "value equal to both min and max",
			value:          time.Second,
			min:            time.Second,
			max:            time.Second,
			wantInRange:    success,
			wantNotInRange: failure,
		},
		{
			name:           "value greater than min and equal to max",
			value:          time.Second,
			min:            time.Second - 1,
			max:            time.Second,
			wantInRange:    success,
			wantNotInRange: failure,
		},
		{
			name:           "value equal to min and smaller than max",
			value:          time.Second,
			min:            time.Second,
			max:            time.Second + 1,
			wantInRange:    success,
			wantNotInRange: failure,
		},
		{
			name:           "value smaller than min",
			value:          time.Second,
			min:            time.Second + 1,
			max:            time.Second + 2,
			wantInRange:    failure,
			wantNotInRange: success,
		},
		{
			name:           "value greater than max",
			value:          time.Second,
			min:            time.Second - 2,
			max:            time.Second - 1,
			wantInRange:    failure,
			wantNotInRange: success,
		},
		{
			name:           "min smaller than max",
			value:          time.Second,
			min:            time.Second + 1,
			max:            time.Second - 1,
			wantInRange:    failure,
			wantNotInRange: success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewDuration(reporter, tc.value).InRange(tc.min, tc.max).
				chain.assert(t, tc.wantInRange)

			NewDuration(reporter, tc.value).NotInRange(tc.min, tc.max).
				chain.assert(t, tc.wantNotInRange)

		})
	}
}

func TestDuration_InList(t *testing.T) {
	cases := []struct {
		name          string
		value         time.Duration
		list          []time.Duration
		wantInList    chainResult
		wantNotInList chainResult
	}{
		{
			name:          "empty list",
			value:         time.Second,
			list:          []time.Duration{},
			wantInList:    failure,
			wantNotInList: failure,
		},
		{
			name:          "value present in list",
			value:         time.Second,
			list:          []time.Duration{time.Second, time.Minute},
			wantInList:    success,
			wantNotInList: failure,
		},
		{
			name:          "value not present in list",
			value:         time.Second,
			list:          []time.Duration{time.Second - 1, time.Second + 1},
			wantInList:    failure,
			wantNotInList: success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewDuration(reporter, tc.value).InList(tc.list...).
				chain.assert(t, tc.wantInList)

			NewDuration(reporter, tc.value).NotInList(tc.list...).
				chain.assert(t, tc.wantNotInList)
		})
	}
}
