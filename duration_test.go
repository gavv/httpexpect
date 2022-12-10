package httpexpect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDurationFailed(t *testing.T) {
	chain := newMockChain(t)
	chain.fail(mockFailure())

	tm := time.Second
	value := newDuration(chain, &tm)

	value.Equal(tm)
	value.NotEqual(tm)
	value.Gt(tm)
	value.Ge(tm)
	value.Lt(tm)
	value.Le(tm)
	value.InRange(tm, tm)
	value.NotInRange(tm, tm)
}

func TestDurationSet(t *testing.T) {
	chain := newMockChain(t)

	tm := time.Second
	value := newDuration(chain, &tm)

	value.IsSet()
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotSet()
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestDurationUnset(t *testing.T) {
	chain := newMockChain(t)

	value := newDuration(chain, nil)

	value.IsSet()
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotSet()
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestDurationEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	assert.Equal(t, time.Second, value.Raw())

	value.Equal(time.Second)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal(time.Minute)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(time.Minute)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(time.Second)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestDurationGreater(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	value.Gt(time.Second - 1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Gt(time.Second)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Ge(time.Second - 1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Ge(time.Second)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Ge(time.Second + 1)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestDurationLesser(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	value.Lt(time.Second + 1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Lt(time.Second)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Le(time.Second + 1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Le(time.Second)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Le(time.Second - 1)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestDurationInRange(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	value.InRange(time.Second, time.Second)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotInRange(time.Second, time.Second)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.InRange(time.Second-1, time.Second)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotInRange(time.Second-1, time.Second)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.InRange(time.Second, time.Second+1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotInRange(time.Second, time.Second+1)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.InRange(time.Second+1, time.Second+2)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotInRange(time.Second+1, time.Second+2)
	value.chain.assertOK(t)
	value.chain.reset()

	value.InRange(time.Second-2, time.Second-1)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotInRange(time.Second-2, time.Second-1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.InRange(time.Second+1, time.Second-1)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotInRange(time.Second+1, time.Second-1)
	value.chain.assertOK(t)
	value.chain.reset()
}
