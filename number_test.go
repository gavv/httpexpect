package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNumberFailed(t *testing.T) {
	chain := makeChain(newMockReporter(t))

	chain.fail("fail")

	value := &Number{chain, 0}

	value.chain.assertFailed(t)

	value.Equal(0)
	value.NotEqual(0)
	value.Gt(0)
	value.Ge(0)
	value.Lt(0)
	value.Le(0)
	value.InRange(0, 0)
}

func TestNumberEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	assert.Equal(t, 1234, int(value.Raw()))

	value.Equal(1234)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal(4321)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(4321)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(1234)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestNumberGreater(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.Gt(1234 - 1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Gt(1234)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Ge(1234 - 1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Ge(1234)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Ge(1234 + 1)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestNumberLesser(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.Lt(1234 + 1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Lt(1234)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Le(1234 + 1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Le(1234)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Le(1234 - 1)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestNumberInRange(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.InRange(1234, 1234)
	value.chain.assertOK(t)
	value.chain.reset()

	value.InRange(1234-1, 1234)
	value.chain.assertOK(t)
	value.chain.reset()

	value.InRange(1234, 1234+1)
	value.chain.assertOK(t)
	value.chain.reset()

	value.InRange(1234+1, 1234+2)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.InRange(1234-2, 1234-1)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.InRange(1234+1, 1234-1)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestNumberConvertEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.Equal(int64(1234))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal(float32(1234))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal("1234")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(int64(4321))
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(float32(4321))
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual("4321")
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestNumberConvertGreater(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.Gt(int64(1233))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Gt(float32(1233))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Gt("1233")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Ge(int64(1233))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Ge(float32(1233))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Ge("1233")
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestNumberConvertLesser(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.Lt(int64(1235))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Lt(float32(1235))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Lt("1235")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Le(int64(1235))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Le(float32(1235))
	value.chain.assertOK(t)
	value.chain.reset()

	value.Le("1235")
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestNumberConvertInRange(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.InRange(int64(1233), float32(1235))
	value.chain.assertOK(t)
	value.chain.reset()

	value.InRange(int64(1233), "1235")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.InRange(nil, 1235)
	value.chain.assertFailed(t)
	value.chain.reset()
}
