package httpexpect

import (
	"testing"
	"time"
	"unsafe"

	"github.com/stretchr/testify/assert"
)

func TestDateTime_Failed(t *testing.T) {
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
		assert.NotEqual(t, unsafe.Pointer(&(value.chain)), unsafe.Pointer(&chain))
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
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
