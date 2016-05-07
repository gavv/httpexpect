package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestValueFailed(t *testing.T) {
	checker := newMockChecker(t)

	checker.Fail("fail")

	value := NewValue(checker, nil)

	value.Object()
	value.Array()
	value.String()
	value.Number()
	value.Boolean()
	value.Null()
	value.NotNull()
}

func TestValueCheckers(t *testing.T) {
	checker := newMockChecker(t)

	var data interface{}

	value := NewValue(checker, data)

	assert.False(t, value.checker == value.Object().checker)
	assert.False(t, value.checker == value.Array().checker)
	assert.False(t, value.checker == value.String().checker)
	assert.False(t, value.checker == value.Number().checker)
	assert.False(t, value.checker == value.Boolean().checker)
}

func TestValueCastNull(t *testing.T) {
	checker := newMockChecker(t)

	var data interface{}

	NewValue(checker, data).Object()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Array()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).String()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Number()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Boolean()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).NotNull()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Null()
	checker.AssertSuccess(t)
	checker.Reset()
}

func TestValueCastIndirectNull(t *testing.T) {
	checker := newMockChecker(t)

	var data []interface{}

	NewValue(checker, data).Object()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Array()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).String()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Number()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Boolean()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).NotNull()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Null()
	checker.AssertSuccess(t)
	checker.Reset()
}

func TestValueCastBad(t *testing.T) {
	checker := newMockChecker(t)

	data := func() {}

	NewValue(checker, data).Object()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Array()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).String()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Number()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Boolean()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).NotNull()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Null()
	checker.AssertFailed(t)
	checker.Reset()
}

func TestValueCastObject(t *testing.T) {
	checker := newMockChecker(t)

	data := map[string]interface{}{}

	NewValue(checker, data).Object()
	checker.AssertSuccess(t)
	checker.Reset()

	NewValue(checker, data).Array()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).String()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Number()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Boolean()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).NotNull()
	checker.AssertSuccess(t)
	checker.Reset()

	NewValue(checker, data).Null()
	checker.AssertFailed(t)
	checker.Reset()
}

func TestValueCastArray(t *testing.T) {
	checker := newMockChecker(t)

	data := []interface{}{}

	NewValue(checker, data).Object()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Array()
	checker.AssertSuccess(t)
	checker.Reset()

	NewValue(checker, data).String()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Number()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Boolean()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).NotNull()
	checker.AssertSuccess(t)
	checker.Reset()

	NewValue(checker, data).Null()
	checker.AssertFailed(t)
	checker.Reset()
}

func TestValueCastString(t *testing.T) {
	checker := newMockChecker(t)

	data := ""

	NewValue(checker, data).Object()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Array()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).String()
	checker.AssertSuccess(t)
	checker.Reset()

	NewValue(checker, data).Number()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Boolean()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).NotNull()
	checker.AssertSuccess(t)
	checker.Reset()

	NewValue(checker, data).Null()
	checker.AssertFailed(t)
	checker.Reset()
}

func TestValueCastNumber(t *testing.T) {
	checker := newMockChecker(t)

	data := 0.0

	NewValue(checker, data).Object()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Array()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).String()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Number()
	checker.AssertSuccess(t)
	checker.Reset()

	NewValue(checker, data).Boolean()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).NotNull()
	checker.AssertSuccess(t)
	checker.Reset()

	NewValue(checker, data).Null()
	checker.AssertFailed(t)
	checker.Reset()
}

func TestValueCastBoolean(t *testing.T) {
	checker := newMockChecker(t)

	data := false

	NewValue(checker, data).Object()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Array()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).String()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Number()
	checker.AssertFailed(t)
	checker.Reset()

	NewValue(checker, data).Boolean()
	checker.AssertSuccess(t)
	checker.Reset()

	NewValue(checker, data).NotNull()
	checker.AssertSuccess(t)
	checker.Reset()

	NewValue(checker, data).Null()
	checker.AssertFailed(t)
	checker.Reset()
}

func TestValueGetObject(t *testing.T) {
	type (
		myMap map[string]interface{}
	)

	checker := newMockChecker(t)

	data1 := map[string]interface{}{"foo": 123.0}

	value1 := NewValue(checker, data1)
	inner1 := value1.Object()

	checker.AssertSuccess(t)
	checker.Reset()
	assert.Equal(t, data1, inner1.Raw())

	data2 := myMap{"foo": 123.0}

	value2 := NewValue(checker, data2)
	inner2 := value2.Object()

	checker.AssertSuccess(t)
	checker.Reset()
	assert.Equal(t, map[string]interface{}(data2), inner2.Raw())
}

func TestValueGetArray(t *testing.T) {
	type (
		myArray []interface{}
	)

	checker := newMockChecker(t)

	data1 := []interface{}{"foo", 123.0}

	value1 := NewValue(checker, data1)
	inner1 := value1.Array()

	checker.AssertSuccess(t)
	checker.Reset()
	assert.Equal(t, data1, inner1.Raw())

	data2 := myArray{"foo", 123.0}

	value2 := NewValue(checker, data2)
	inner2 := value2.Array()

	checker.AssertSuccess(t)
	checker.Reset()
	assert.Equal(t, []interface{}(data2), inner2.Raw())
}

func TestValueGetString(t *testing.T) {
	checker := newMockChecker(t)

	value := NewValue(checker, "foo")
	inner := value.String()

	checker.AssertSuccess(t)
	checker.Reset()
	assert.Equal(t, "foo", inner.Raw())
}

func TestValueGetNumber(t *testing.T) {
	type (
		myInt int
	)

	checker := newMockChecker(t)

	data1 := 123.0

	value1 := NewValue(checker, data1)
	inner1 := value1.Number()

	checker.AssertSuccess(t)
	checker.Reset()
	assert.Equal(t, data1, inner1.Raw())

	data2 := 123

	value2 := NewValue(checker, data2)
	inner2 := value2.Number()

	checker.AssertSuccess(t)
	checker.Reset()
	assert.Equal(t, float64(data2), inner2.Raw())

	data3 := myInt(123)

	value3 := NewValue(checker, data3)
	inner3 := value3.Number()

	checker.AssertSuccess(t)
	checker.Reset()
	assert.Equal(t, float64(data3), inner3.Raw())
}

func TestValueGetBoolean(t *testing.T) {
	checker := newMockChecker(t)

	value1 := NewValue(checker, true)
	inner1 := value1.Boolean()

	checker.AssertSuccess(t)
	checker.Reset()
	assert.Equal(t, true, inner1.Raw())

	value2 := NewValue(checker, false)
	inner2 := value2.Boolean()

	checker.AssertSuccess(t)
	checker.Reset()
	assert.Equal(t, false, inner2.Raw())
}
