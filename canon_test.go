package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCanonNumber(t *testing.T) {
	type (
		myInt int
	)

	checker := newMockChecker(t)

	d1, ok := canonNumber(checker, 123)
	assert.True(t, ok)
	assert.Equal(t, 123.0, d1)
	checker.AssertSuccess(t)
	checker.Reset()

	d2, ok := canonNumber(checker, 123.0)
	assert.True(t, ok)
	assert.Equal(t, 123.0, d2)
	checker.AssertSuccess(t)
	checker.Reset()

	d3, ok := canonNumber(checker, myInt(123))
	assert.True(t, ok)
	assert.Equal(t, 123.0, d3)
	checker.AssertSuccess(t)
	checker.Reset()

	_, ok = canonNumber(checker, "123")
	assert.False(t, ok)
	checker.AssertFailed(t)
	checker.Reset()

	_, ok = canonNumber(checker, nil)
	assert.False(t, ok)
	checker.AssertFailed(t)
	checker.Reset()
}

func TestCanonArray(t *testing.T) {
	type (
		myArray []interface{}
		myInt   int
	)

	checker := newMockChecker(t)

	d1, ok := canonArray(checker, []interface{}{123.0, 456.0})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{123.0, 456.0}, d1)
	checker.AssertSuccess(t)
	checker.Reset()

	d2, ok := canonArray(checker, myArray{myInt(123), 456.0})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{123.0, 456.0}, d2)
	checker.AssertSuccess(t)
	checker.Reset()

	_, ok = canonArray(checker, "123")
	assert.False(t, ok)
	checker.AssertFailed(t)
	checker.Reset()

	_, ok = canonArray(checker, func() {})
	assert.False(t, ok)
	checker.AssertFailed(t)
	checker.Reset()

	_, ok = canonArray(checker, nil)
	assert.False(t, ok)
	checker.AssertFailed(t)
	checker.Reset()

	_, ok = canonArray(checker, []interface{}(nil))
	assert.False(t, ok)
	checker.AssertFailed(t)
	checker.Reset()
}

func TestCanonMap(t *testing.T) {
	type (
		myMap map[string]interface{}
		myInt int
	)

	checker := newMockChecker(t)

	d1, ok := canonMap(checker, map[string]interface{}{"foo": 123.0})
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{"foo": 123.0}, d1)
	checker.AssertSuccess(t)
	checker.Reset()

	d2, ok := canonMap(checker, myMap{"foo": myInt(123)})
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{"foo": 123.0}, d2)
	checker.AssertSuccess(t)
	checker.Reset()

	_, ok = canonMap(checker, "123")
	assert.False(t, ok)
	checker.AssertFailed(t)
	checker.Reset()

	_, ok = canonMap(checker, func() {})
	assert.False(t, ok)
	checker.AssertFailed(t)
	checker.Reset()

	_, ok = canonMap(checker, nil)
	assert.False(t, ok)
	checker.AssertFailed(t)
	checker.Reset()

	_, ok = canonMap(checker, map[string]interface{}(nil))
	assert.False(t, ok)
	checker.AssertFailed(t)
	checker.Reset()
}
