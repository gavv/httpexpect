package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCanonNumber(t *testing.T) {
	type (
		myInt int
	)

	checker := &mockChecker{}

	d1, ok := canonNumber(checker, 123)
	assert.True(t, ok)
	assert.Equal(t, 123.0, d1)

	d2, ok := canonNumber(checker, 123.0)
	assert.True(t, ok)
	assert.Equal(t, 123.0, d2)

	d3, ok := canonNumber(checker, myInt(123))
	assert.True(t, ok)
	assert.Equal(t, 123.0, d3)

	_, ok = canonNumber(checker, "123")
	assert.False(t, ok)

	_, ok = canonNumber(checker, nil)
	assert.False(t, ok)
}

func TestCanonArray(t *testing.T) {
	type (
		myArray []interface{}
		myInt   int
	)

	checker := &mockChecker{}

	d1, ok := canonArray(checker, []interface{}{123.0, 456.0})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{123.0, 456.0}, d1)

	d2, ok := canonArray(checker, myArray{myInt(123), 456.0})
	assert.True(t, ok)
	assert.Equal(t, []interface{}{123.0, 456.0}, d2)

	d3, ok := canonArray(checker, nil)
	assert.True(t, ok)
	assert.Equal(t, []interface{}(nil), d3)
}

func TestCanonMap(t *testing.T) {
	type (
		myMap map[string]interface{}
		myInt int
	)

	checker := &mockChecker{}

	d1, ok := canonMap(checker, map[string]interface{}{"foo": 123.0})
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{"foo": 123.0}, d1)

	d2, ok := canonMap(checker, myMap{"foo": myInt(123)})
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}{"foo": 123.0}, d2)

	d3, ok := canonMap(checker, nil)
	assert.True(t, ok)
	assert.Equal(t, map[string]interface{}(nil), d3)
}
