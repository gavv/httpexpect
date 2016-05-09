package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestArrayFailed(t *testing.T) {
	checker := newMockChecker(t)

	checker.Fail("fail")

	value := NewArray(checker, nil)

	assert.False(t, value.Length() == nil)
	assert.False(t, value.Element(0) == nil)

	value.Empty()
	value.NotEmpty()
	value.Equal(nil)
	value.NotEqual(nil)
	value.Elements("foo")
	value.Contains("foo")
	value.NotContains("foo")
	value.ContainsOnly("foo")
}

func TestArrayGetters(t *testing.T) {
	checker := newMockChecker(t)

	value := NewArray(checker, []interface{}{"foo", 123.0})

	assert.Equal(t, 2.0, value.Length().Raw())

	assert.Equal(t, "foo", value.Element(0).Raw().(string))
	assert.Equal(t, 123.0, value.Element(1).Raw().(float64))
	checker.AssertSuccess(t)
	checker.Reset()

	assert.Equal(t, nil, value.Element(2).Raw())
	checker.AssertFailed(t)
	checker.Reset()

	assert.False(t, value.checker == value.Length().checker)
	assert.False(t, value.checker == value.Element(0).checker)
}

func TestArrayEmpty(t *testing.T) {
	checker := newMockChecker(t)

	value1 := NewArray(checker, nil)

	_ = value1
	checker.AssertFailed(t)
	checker.Reset()

	value2 := NewArray(checker, []interface{}{})

	value2.Empty()
	checker.AssertSuccess(t)
	checker.Reset()

	value2.NotEmpty()
	checker.AssertFailed(t)
	checker.Reset()

	value3 := NewArray(checker, []interface{}{""})

	value3.Empty()
	checker.AssertFailed(t)
	checker.Reset()

	value3.NotEmpty()
	checker.AssertSuccess(t)
	checker.Reset()
}

func TestArrayEqualEmpty(t *testing.T) {
	checker := newMockChecker(t)

	value := NewArray(checker, []interface{}{})

	assert.Equal(t, []interface{}{}, value.Raw())

	value.Equal([]interface{}{})
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual([]interface{}{})
	checker.AssertFailed(t)
	checker.Reset()

	value.Equal([]interface{}{""})
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual([]interface{}{""})
	checker.AssertSuccess(t)
	checker.Reset()
}

func TestArrayEqualNotEmpty(t *testing.T) {
	checker := newMockChecker(t)

	value := NewArray(checker, []interface{}{"foo", "bar"})

	assert.Equal(t, []interface{}{"foo", "bar"}, value.Raw())

	value.Equal([]interface{}{})
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual([]interface{}{})
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal([]interface{}{"foo"})
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual([]interface{}{"foo"})
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal([]interface{}{"bar", "foo"})
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual([]interface{}{"bar", "foo"})
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal([]interface{}{"foo", "bar"})
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual([]interface{}{"foo", "bar"})
	checker.AssertFailed(t)
	checker.Reset()
}

func TestArrayElements(t *testing.T) {
	checker := newMockChecker(t)

	value := NewArray(checker, []interface{}{123, "foo"})

	value.Elements(123)
	checker.AssertFailed(t)
	checker.Reset()

	value.Elements("foo")
	checker.AssertFailed(t)
	checker.Reset()

	value.Elements("foo", 123)
	checker.AssertFailed(t)
	checker.Reset()

	value.Elements(123, "foo", "foo")
	checker.AssertFailed(t)
	checker.Reset()

	value.Elements(123, "foo")
	checker.AssertSuccess(t)
	checker.Reset()
}

func TestArrayContains(t *testing.T) {
	checker := newMockChecker(t)

	value := NewArray(checker, []interface{}{123, "foo"})

	value.Contains(123)
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotContains(123)
	checker.AssertFailed(t)
	checker.Reset()

	value.Contains("foo", 123)
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotContains("foo", 123)
	checker.AssertFailed(t)
	checker.Reset()

	value.Contains("foo", "foo")
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotContains("foo", "foo")
	checker.AssertFailed(t)
	checker.Reset()

	value.Contains(123, "foo", "FOO")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContains(123, "foo", "FOO")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContains("FOO")
	checker.AssertSuccess(t)
	checker.Reset()

	value.Contains([]interface{}{123, "foo"})
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContains([]interface{}{123, "foo"})
	checker.AssertSuccess(t)
	checker.Reset()
}

func TestArrayContainsOnly(t *testing.T) {
	checker := newMockChecker(t)

	value := NewArray(checker, []interface{}{123, "foo"})

	value.ContainsOnly(123)
	checker.AssertFailed(t)
	checker.Reset()

	value.ContainsOnly("foo")
	checker.AssertFailed(t)
	checker.Reset()

	value.ContainsOnly(123, "foo", "foo")
	checker.AssertFailed(t)
	checker.Reset()

	value.ContainsOnly(123, "foo")
	checker.AssertSuccess(t)
	checker.Reset()

	value.ContainsOnly("foo", 123)
	checker.AssertSuccess(t)
	checker.Reset()
}

func TestArrayConvertEqual(t *testing.T) {
	type (
		myArray []interface{}
		myInt   int
	)

	checker := newMockChecker(t)

	value := NewArray(checker, []interface{}{123, 456})

	assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

	value.Equal(myArray{myInt(123), 456.0})
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual(myArray{myInt(123), 456.0})
	checker.AssertFailed(t)
	checker.Reset()

	value.Equal([]interface{}{"123", "456"})
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual([]interface{}{"123", "456"})
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal(nil)
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual(nil)
	checker.AssertFailed(t)
	checker.Reset()
}

func TestArrayConvertElements(t *testing.T) {
	type (
		myInt int
	)

	checker := newMockChecker(t)

	value := NewArray(checker, []interface{}{123, 456})

	assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

	value.Elements(myInt(123), 456.0)
	checker.AssertSuccess(t)
	checker.Reset()

	value.Elements(func() {})
	checker.AssertFailed(t)
	checker.Reset()
}

func TestArrayConvertContains(t *testing.T) {
	type (
		myInt int
	)

	checker := newMockChecker(t)

	value := NewArray(checker, []interface{}{123, 456})

	assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

	value.Contains(myInt(123), 456.0)
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotContains(myInt(123), 456.0)
	checker.AssertFailed(t)
	checker.Reset()

	value.ContainsOnly(myInt(123), 456.0)
	checker.AssertSuccess(t)
	checker.Reset()

	value.Contains("123")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContains("123")
	checker.AssertSuccess(t)
	checker.Reset()

	value.ContainsOnly("123.0", "456.0")
	checker.AssertFailed(t)
	checker.Reset()

	value.Contains(func() {})
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContains(func() {})
	checker.AssertFailed(t)
	checker.Reset()

	value.ContainsOnly(func() {})
	checker.AssertFailed(t)
	checker.Reset()
}
