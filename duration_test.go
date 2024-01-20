package httpexpect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDuration_FailedChain(t *testing.T) {
	chain := newMockChain(t, flagFailed)

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

func TestDuration_Raw(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	assert.Equal(t, time.Second, value.Raw())
	value.chain.assert(t, success)
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
	cases := []struct {
		name      string
		duration  time.Duration
		value     time.Duration
		wantEqual chainResult
	}{
		{
			name:      "compare equivalent durations",
			duration:  time.Second,
			value:     time.Second,
			wantEqual: success,
		},
		{
			name:      "compare non-equivalent durations",
			duration:  time.Second,
			value:     time.Minute,
			wantEqual: failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewDuration(reporter, tc.duration).IsEqual(tc.value).
				chain.assert(t, tc.wantEqual)

			NewDuration(reporter, tc.duration).NotEqual(tc.value).
				chain.assert(t, !tc.wantEqual)
		})
	}
}

func TestDuration_IsGreater(t *testing.T) {
	cases := []struct {
		name     string
		duration time.Duration
		value    time.Duration
		wantGt   chainResult
		wantGe   chainResult
	}{
		{
			name:     "duration is lesser",
			duration: time.Second,
			value:    time.Second + 1,
			wantGt:   failure,
			wantGe:   failure,
		},
		{
			name:     "duration is equal",
			duration: time.Second,
			value:    time.Second,
			wantGt:   failure,
			wantGe:   success,
		},
		{
			name:     "duration is greater",
			duration: time.Second,
			value:    time.Second - 1,
			wantGt:   success,
			wantGe:   success,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewDuration(reporter, tc.duration).Gt(tc.value).
				chain.assert(t, tc.wantGt)

			NewDuration(reporter, tc.duration).Ge(tc.value).
				chain.assert(t, tc.wantGe)
		})
	}
}

func TestDuration_IsLesser(t *testing.T) {
	cases := []struct {
		name     string
		duration time.Duration
		value    time.Duration
		wantLt   chainResult
		wantLe   chainResult
	}{
		{
			name:     "duration is lesser",
			duration: time.Second,
			value:    time.Second + 1,
			wantLt:   success,
			wantLe:   success,
		},
		{
			name:     "duration is equal",
			duration: time.Second,
			value:    time.Second,
			wantLt:   failure,
			wantLe:   success,
		},
		{
			name:     "duration is greater",
			duration: time.Second,
			value:    time.Second - 1,
			wantLt:   failure,
			wantLe:   failure,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			NewDuration(reporter, tc.duration).Lt(tc.value).
				chain.assert(t, tc.wantLt)

			NewDuration(reporter, tc.duration).Le(tc.value).
				chain.assert(t, tc.wantLe)
		})
	}
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
