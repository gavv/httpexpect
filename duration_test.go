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
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotSet()
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDurationUnset(t *testing.T) {
	chain := newMockChain(t)

	value := newDuration(chain, nil)

	value.IsSet()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotSet()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestDurationEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	assert.Equal(t, time.Second, value.Raw())

	value.Equal(time.Second)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Equal(time.Minute)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(time.Minute)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(time.Second)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDurationGreater(t *testing.T) {
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

func TestDurationLesser(t *testing.T) {
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

func TestDurationInRange(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	value.InRange(time.Second, time.Second)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Second, time.Second)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Second-1, time.Second)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Second-1, time.Second)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Second, time.Second+1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Second, time.Second+1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Second+1, time.Second+2)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Second+1, time.Second+2)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Second-2, time.Second-1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Second-2, time.Second-1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Second+1, time.Second-1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Second+1, time.Second-1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}
