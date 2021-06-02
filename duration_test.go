package httpexpect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDurationFailed(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	chain.fail("fail")

	ts := time.Second

	value := &Duration{chain, &ts}

	value.chain.assertFailed(t)

	value.Equal(ts)
	value.NotEqual(ts)
	value.Gt(ts)
	value.Ge(ts)
	value.Lt(ts)
	value.Le(ts)
	value.InRange(ts, ts)
}

func TestDurationNil(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	ts := time.Second

	value := &Duration{chain, nil}

	value.chain.assertOK(t)

	value.Equal(ts)
	value.NotEqual(ts)
	value.Gt(ts)
	value.Ge(ts)
	value.Lt(ts)
	value.Le(ts)
	value.InRange(ts, ts)
}

func TestDurationSet(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	ts := time.Second

	value := &Duration{chain, &ts}

	value.IsSet()
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotSet()
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestDurationUnset(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	value := &Duration{chain, nil}

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

	value.InRange(time.Second-1, time.Second)
	value.chain.assertOK(t)
	value.chain.reset()

	value.InRange(time.Second, time.Second+1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.InRange(time.Second+1, time.Second+2)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.InRange(time.Second-2, time.Second-1)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.InRange(time.Second+1, time.Second-1)
	value.chain.assertFailed(t)
	value.chain.reset()
}
