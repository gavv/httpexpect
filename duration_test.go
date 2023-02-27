package httpexpect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDuration_FailedChain(t *testing.T) {
	chain := newMockChain(t)
	chain.setFailed()

	tm := time.Second
	value := newDuration(chain, &tm)
	value.chain.assertFailed(t)

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
		value.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewDurationC(Config{
			Reporter: reporter,
		}, tm)
		value.IsEqual(tm)
		value.chain.assertNotFailed(t)
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

func TestDuration_Set(t *testing.T) {
	chain := newMockChain(t)

	tm := time.Second
	value := newDuration(chain, &tm)

	value.IsSet()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotSet()
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDuration_Unset(t *testing.T) {
	chain := newMockChain(t)

	value := newDuration(chain, nil)

	value.IsSet()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotSet()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestDuration_IsEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	assert.Equal(t, time.Second, value.Raw())

	value.IsEqual(time.Second)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsEqual(time.Minute)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(time.Minute)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(time.Second)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDuration_IsGreater(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	value.Gt(time.Second - 1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Gt(time.Second)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Ge(time.Second - 1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Ge(time.Second)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Ge(time.Second + 1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDuration_IsLesser(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	value.Lt(time.Second + 1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Lt(time.Second)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Le(time.Second + 1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Le(time.Second)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Le(time.Second - 1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDuration_InRange(t *testing.T) {
	cases := map[string]struct {
		value            time.Duration
		min              time.Duration
		max              time.Duration
		expectInRange    bool
		expectNotInRange bool
	}{
		"value equal to both min and max": {
			value:            time.Second,
			min:              time.Second,
			max:              time.Second,
			expectInRange:    true,
			expectNotInRange: false,
		},
		"value greater than min and equal to max": {
			value:            time.Second,
			min:              time.Second - 1,
			max:              time.Second,
			expectInRange:    true,
			expectNotInRange: false,
		},
		"value equal to min and smaller than max": {
			value:            time.Second,
			min:              time.Second,
			max:              time.Second + 1,
			expectInRange:    true,
			expectNotInRange: false,
		},
		"value smaller than min": {
			value:            time.Second,
			min:              time.Second + 1,
			max:              time.Second + 2,
			expectInRange:    false,
			expectNotInRange: true,
		},
		"value greater than max": {
			value:            time.Second,
			min:              time.Second - 2,
			max:              time.Second - 1,
			expectInRange:    false,
			expectNotInRange: true,
		},
		"min smaller than max": {
			value:            time.Second,
			min:              time.Second + 1,
			max:              time.Second - 1,
			expectInRange:    false,
			expectNotInRange: true,
		},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if instance.expectInRange {
				NewDuration(reporter, instance.value).
					InRange(instance.min, instance.max).
					chain.assertNotFailed(t)
			} else {
				NewDuration(reporter, instance.value).
					InRange(instance.min, instance.max).
					chain.assertFailed(t)
			}
			if instance.expectNotInRange {
				NewDuration(reporter, instance.value).
					NotInRange(instance.min, instance.max).
					chain.assertNotFailed(t)
			} else {
				NewDuration(reporter, instance.value).
					NotInRange(instance.min, instance.max).
					chain.assertFailed(t)
			}
		})
	}
}

func TestDuration_InList(t *testing.T) {
	cases := map[string]struct {
		value           time.Duration
		list            []time.Duration
		expectInList    bool
		expectNotInList bool
	}{
		"empty list": {
			value:           time.Second,
			list:            []time.Duration{},
			expectInList:    false,
			expectNotInList: false,
		},
		"value present in list": {
			value:           time.Second,
			list:            []time.Duration{time.Second, time.Minute},
			expectInList:    true,
			expectNotInList: false,
		},
		"value not present in list": {
			value:           time.Second,
			list:            []time.Duration{time.Second - 1, time.Second + 1},
			expectInList:    false,
			expectNotInList: true,
		},
	}

	for name, instance := range cases {
		t.Run(name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if instance.expectInList {
				NewDuration(reporter, instance.value).
					InList(instance.list...).
					chain.assertNotFailed(t)
			} else {
				NewDuration(reporter, instance.value).
					InList(instance.list...).
					chain.assertFailed(t)
			}
			if instance.expectNotInList {
				NewDuration(reporter, instance.value).
					NotInList(instance.list...).
					chain.assertNotFailed(t)
			} else {
				NewDuration(reporter, instance.value).
					NotInList(instance.list...).
					chain.assertFailed(t)
			}
		})
	}
}
