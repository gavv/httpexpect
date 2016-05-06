package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNumberEqual(t *testing.T) {
	checker := newMockChecker(t)

	value := NewNumber(checker, 1234)

	assert.Equal(t, 1234, int(value.Raw()))

	value.Equal(1234)
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal(4321)
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual(4321)
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual(1234)
	checker.AssertFailed(t)
	checker.Reset()
}

func TestNumberGreater(t *testing.T) {
	checker := newMockChecker(t)

	value := NewNumber(checker, 1234)

	value.Gt(1234 - 1)
	checker.AssertSuccess(t)
	checker.Reset()

	value.Gt(1234)
	checker.AssertFailed(t)
	checker.Reset()

	value.Ge(1234 - 1)
	checker.AssertSuccess(t)
	checker.Reset()

	value.Ge(1234)
	checker.AssertSuccess(t)
	checker.Reset()

	value.Ge(1234 + 1)
	checker.AssertFailed(t)
	checker.Reset()
}

func TestNumberLesser(t *testing.T) {
	checker := newMockChecker(t)

	value := NewNumber(checker, 1234)

	value.Lt(1234 + 1)
	checker.AssertSuccess(t)
	checker.Reset()

	value.Lt(1234)
	checker.AssertFailed(t)
	checker.Reset()

	value.Le(1234 + 1)
	checker.AssertSuccess(t)
	checker.Reset()

	value.Le(1234)
	checker.AssertSuccess(t)
	checker.Reset()

	value.Le(1234 - 1)
	checker.AssertFailed(t)
	checker.Reset()
}

func TestNumberInRange(t *testing.T) {
	checker := newMockChecker(t)

	value := NewNumber(checker, 1234)

	value.InRange(1234, 1234)
	checker.AssertSuccess(t)
	checker.Reset()

	value.InRange(1234 - 1, 1234)
	checker.AssertSuccess(t)
	checker.Reset()

	value.InRange(1234, 1234 + 1)
	checker.AssertSuccess(t)
	checker.Reset()

	value.InRange(1234 + 1, 1234 + 2)
	checker.AssertFailed(t)
	checker.Reset()

	value.InRange(1234 - 2, 1234 - 1)
	checker.AssertFailed(t)
	checker.Reset()

	value.InRange(1234 + 1, 1234 - 1)
	checker.AssertFailed(t)
	checker.Reset()
}

func TestNumberConvertEqual(t *testing.T) {
	checker := newMockChecker(t)

	value := NewNumber(checker, 1234)

	value.Equal(int64(1234))
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal(float32(1234))
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal("1234")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual(int64(4321))
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual(float32(4321))
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual("4321")
	checker.AssertFailed(t)
	checker.Reset()
}

func TestNumberConvertGreater(t *testing.T) {
	checker := newMockChecker(t)

	value := NewNumber(checker, 1234)

	value.Gt(int64(1233))
	checker.AssertSuccess(t)
	checker.Reset()

	value.Gt(float32(1233))
	checker.AssertSuccess(t)
	checker.Reset()

	value.Gt("1233")
	checker.AssertFailed(t)
	checker.Reset()

	value.Ge(int64(1233))
	checker.AssertSuccess(t)
	checker.Reset()

	value.Ge(float32(1233))
	checker.AssertSuccess(t)
	checker.Reset()

	value.Ge("1233")
	checker.AssertFailed(t)
	checker.Reset()
}

func TestNumberConvertLesser(t *testing.T) {
	checker := newMockChecker(t)

	value := NewNumber(checker, 1234)

	value.Lt(int64(1235))
	checker.AssertSuccess(t)
	checker.Reset()

	value.Lt(float32(1235))
	checker.AssertSuccess(t)
	checker.Reset()

	value.Lt("1235")
	checker.AssertFailed(t)
	checker.Reset()

	value.Le(int64(1235))
	checker.AssertSuccess(t)
	checker.Reset()

	value.Le(float32(1235))
	checker.AssertSuccess(t)
	checker.Reset()

	value.Le("1235")
	checker.AssertFailed(t)
	checker.Reset()
}

func TestNumberConvertInRange(t *testing.T) {
	checker := newMockChecker(t)

	value := NewNumber(checker, 1234)

	value.InRange(int64(1233), float32(1235))
	checker.AssertSuccess(t)
	checker.Reset()

	value.InRange(int64(1233), "1235")
	checker.AssertFailed(t)
	checker.Reset()
}
