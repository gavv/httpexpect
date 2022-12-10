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
}

func TestDateTimeEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDateTime(reporter, time.Unix(0, 1234))

	assert.True(t, time.Unix(0, 1234).Equal(value.Raw()))

	value.Equal(time.Unix(0, 1234))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal(time.Unix(0, 4321))
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(time.Unix(0, 4321))
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(time.Unix(0, 1234))
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestDateTimeGreater(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDateTime(reporter, time.Unix(0, 1234))

	value.Gt(time.Unix(0, 1234-1))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Gt(time.Unix(0, 1234))
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Ge(time.Unix(0, 1234-1))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Ge(time.Unix(0, 1234))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Ge(time.Unix(0, 1234+1))
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestDateTimeLesser(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDateTime(reporter, time.Unix(0, 1234))

	value.Lt(time.Unix(0, 1234+1))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Lt(time.Unix(0, 1234))
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Le(time.Unix(0, 1234+1))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Le(time.Unix(0, 1234))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Le(time.Unix(0, 1234-1))
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestDateTimeInRange(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDateTime(reporter, time.Unix(0, 1234))

	value.InRange(time.Unix(0, 1234), time.Unix(0, 1234))
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotInRange(time.Unix(0, 1234), time.Unix(0, 1234))
	value.chain.assertFailed(t)
	value.chain.reset()

	value.InRange(time.Unix(0, 1234-1), time.Unix(0, 1234))
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotInRange(time.Unix(0, 1234-1), time.Unix(0, 1234))
	value.chain.assertFailed(t)
	value.chain.reset()

	value.InRange(time.Unix(0, 1234), time.Unix(0, 1234+1))
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotInRange(time.Unix(0, 1234), time.Unix(0, 1234+1))
	value.chain.assertFailed(t)
	value.chain.reset()

	value.InRange(time.Unix(0, 1234+1), time.Unix(0, 1234+2))
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotInRange(time.Unix(0, 1234+1), time.Unix(0, 1234+2))
	value.chain.assertOK(t)
	value.chain.reset()

	value.InRange(time.Unix(0, 1234-2), time.Unix(0, 1234-1))
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotInRange(time.Unix(0, 1234-2), time.Unix(0, 1234-1))
	value.chain.assertOK(t)
	value.chain.reset()

	value.InRange(time.Unix(0, 1234+1), time.Unix(0, 1234-1))
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotInRange(time.Unix(0, 1234+1), time.Unix(0, 1234-1))
	value.chain.assertOK(t)
	value.chain.reset()
}
