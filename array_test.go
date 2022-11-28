package httpexpect

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArrayFailed(t *testing.T) {
	check := func(value *Array) {
		value.chain.assertFailed(t)

		value.Path("$")
		value.Schema("")

		assert.NotNil(t, value.Length())
		assert.NotNil(t, value.Element(0))
		assert.NotNil(t, value.Iter())
		assert.Equal(t, 0, len(value.Iter()))

		value.Length().chain.assertFailed(t)
		value.Element(0).chain.assertFailed(t)
		value.First().chain.assertFailed(t)
		value.Last().chain.assertFailed(t)

		value.Empty()
		value.NotEmpty()
		value.Equal([]interface{}{})
		value.NotEqual([]interface{}{})
		value.EqualUnordered([]interface{}{})
		value.NotEqualUnordered([]interface{}{})
		value.Elements("foo")
		value.NotElements("foo")
		value.Contains("foo")
		value.NotContains("foo")
		value.ContainsOnly("foo")
		value.NotContainsOnly("foo")
		value.ContainsAny("foo")
		value.NotContainsAny("foo")
		value.Every(func(_ int, val *Value) {
			val.String().NotEmpty()
		})
	}

	t.Run("failed_chain", func(t *testing.T) {
		chain := newMockChain(t)
		chain.fail(AssertionFailure{})

		value := newArray(chain, []interface{}{})

		check(value)
	})

	t.Run("nil_value", func(t *testing.T) {
		chain := newMockChain(t)

		value := newArray(chain, nil)

		check(value)
	})

	t.Run("failed_chain_nil_value", func(t *testing.T) {
		chain := newMockChain(t)
		chain.fail(AssertionFailure{})

		value := newArray(chain, nil)

		check(value)
	})
}

func TestArrayGetters(t *testing.T) {
	reporter := newMockReporter(t)

	a := []interface{}{"foo", 123.0}

	value := NewArray(reporter, a)

	assert.Equal(t, a, value.Raw())
	value.chain.assertOK(t)
	value.chain.reset()

	assert.Equal(t, a, value.Path("$").Raw())
	value.chain.assertOK(t)
	value.chain.reset()

	value.Schema(`{"type": "array"}`)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Schema(`{"type": "object"}`)
	value.chain.assertFailed(t)
	value.chain.reset()

	assert.Equal(t, 2.0, value.Length().Raw())

	assert.Equal(t, "foo", value.Element(0).Raw())
	assert.Equal(t, 123.0, value.Element(1).Raw())
	value.chain.assertOK(t)
	value.chain.reset()

	assert.Equal(t, nil, value.Element(2).Raw())
	value.chain.assertFailed(t)
	value.chain.reset()

	it := value.Iter()
	assert.Equal(t, 2, len(it))
	assert.Equal(t, "foo", it[0].Raw())
	assert.Equal(t, 123.0, it[1].Raw())
	value.chain.assertOK(t)
	value.chain.reset()

	assert.Equal(t, "foo", value.First().Raw())
	assert.Equal(t, 123.0, value.Last().Raw())
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestArrayEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewArray(reporter, nil)

	_ = value1
	value1.chain.assertFailed(t)
	value1.chain.reset()

	value2 := NewArray(reporter, []interface{}{})

	value2.Empty()
	value2.chain.assertOK(t)
	value2.chain.reset()

	value2.NotEmpty()
	value2.chain.assertFailed(t)
	value2.chain.reset()

	value3 := NewArray(reporter, []interface{}{""})

	value3.Empty()
	value3.chain.assertFailed(t)
	value3.chain.reset()

	value3.NotEmpty()
	value3.chain.assertOK(t)
	value3.chain.reset()
}

func TestArrayEmptyGetters(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{})

	assert.NotNil(t, value.Element(0))
	value.chain.assertFailed(t)
	value.chain.reset()

	assert.NotNil(t, value.First())
	value.chain.assertFailed(t)
	value.chain.reset()

	assert.NotNil(t, value.Last())
	value.chain.assertFailed(t)
	value.chain.reset()

	assert.NotNil(t, value.Iter())
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestArrayEqualEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{})

	assert.Equal(t, []interface{}{}, value.Raw())

	value.Equal([]interface{}{})
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual([]interface{}{})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Equal([]interface{}{""})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual([]interface{}{""})
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestArrayEqualNotEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{"foo", "bar"})

	assert.Equal(t, []interface{}{"foo", "bar"}, value.Raw())

	value.Equal([]interface{}{})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual([]interface{}{})
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal([]interface{}{"foo"})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual([]interface{}{"foo"})
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal([]interface{}{"bar", "foo"})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual([]interface{}{"bar", "foo"})
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal([]interface{}{"foo", "bar"})
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual([]interface{}{"foo", "bar"})
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestArrayEqualTypes(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewArray(reporter, []interface{}{"foo", "bar"})
	value2 := NewArray(reporter, []interface{}{123, 456})
	value3 := NewArray(reporter, []interface{}{
		map[string]interface{}{
			"foo": 123,
		},
		map[string]interface{}{
			"foo": 456,
		},
	})

	value1.Equal([]string{"foo", "bar"})
	value1.chain.assertOK(t)
	value1.chain.reset()

	value1.Equal([]string{"bar", "foo"})
	value1.chain.assertFailed(t)
	value1.chain.reset()

	value1.NotEqual([]string{"foo", "bar"})
	value1.chain.assertFailed(t)
	value1.chain.reset()

	value1.NotEqual([]string{"bar", "foo"})
	value1.chain.assertOK(t)
	value1.chain.reset()

	value2.Equal([]int{123, 456})
	value2.chain.assertOK(t)
	value2.chain.reset()

	value2.Equal([]int{456, 123})
	value2.chain.assertFailed(t)
	value2.chain.reset()

	value2.NotEqual([]int{123, 456})
	value2.chain.assertFailed(t)
	value2.chain.reset()

	value2.NotEqual([]int{456, 123})
	value2.chain.assertOK(t)
	value2.chain.reset()

	type S struct {
		Foo int `json:"foo"`
	}

	value3.Equal([]S{{123}, {456}})
	value3.chain.assertOK(t)
	value3.chain.reset()

	value3.Equal([]S{{456}, {123}})
	value3.chain.assertFailed(t)
	value3.chain.reset()

	value3.NotEqual([]S{{123}, {456}})
	value3.chain.assertFailed(t)
	value3.chain.reset()

	value3.NotEqual([]S{{456}, {123}})
	value3.chain.assertOK(t)
	value3.chain.reset()
}

func TestArrayEqualUnordered(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("without_duplicates", func(t *testing.T) {
		value := NewArray(reporter, []interface{}{123, "foo"})

		value.EqualUnordered([]interface{}{123})
		value.chain.assertFailed(t)
		value.chain.reset()

		value.NotEqualUnordered([]interface{}{123})
		value.chain.assertOK(t)
		value.chain.reset()

		value.EqualUnordered([]interface{}{"foo"})
		value.chain.assertFailed(t)
		value.chain.reset()

		value.NotEqualUnordered([]interface{}{"foo"})
		value.chain.assertOK(t)
		value.chain.reset()

		value.EqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assertFailed(t)
		value.chain.reset()

		value.NotEqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assertOK(t)
		value.chain.reset()

		value.EqualUnordered([]interface{}{123, "foo"})
		value.chain.assertOK(t)
		value.chain.reset()

		value.NotEqualUnordered([]interface{}{123, "foo"})
		value.chain.assertFailed(t)
		value.chain.reset()

		value.EqualUnordered([]interface{}{"foo", 123})
		value.chain.assertOK(t)
		value.chain.reset()

		value.NotEqualUnordered([]interface{}{"foo", 123})
		value.chain.assertFailed(t)
		value.chain.reset()
	})

	t.Run("with_duplicates", func(t *testing.T) {
		value := NewArray(reporter, []interface{}{123, "foo", "foo"})

		value.EqualUnordered([]interface{}{123, "foo"})
		value.chain.assertFailed(t)
		value.chain.reset()

		value.NotEqualUnordered([]interface{}{123, "foo"})
		value.chain.assertOK(t)
		value.chain.reset()

		value.EqualUnordered([]interface{}{123, 123, "foo"})
		value.chain.assertFailed(t)
		value.chain.reset()

		value.NotEqualUnordered([]interface{}{123, 123, "foo"})
		value.chain.assertOK(t)
		value.chain.reset()

		value.EqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assertOK(t)
		value.chain.reset()

		value.NotEqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assertFailed(t)
		value.chain.reset()

		value.EqualUnordered([]interface{}{"foo", 123, "foo"})
		value.chain.assertOK(t)
		value.chain.reset()

		value.NotEqualUnordered([]interface{}{"foo", 123, "foo"})
		value.chain.assertFailed(t)
		value.chain.reset()
	})
}

func TestArrayElements(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, "foo"})

	value.Elements(123)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Elements("foo")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Elements("foo", 123)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Elements(123, "foo", "foo")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Elements(123, "foo")
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestArrayNotElements(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, "foo"})

	value.NotElements(123)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotElements("foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotElements("foo", 123)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotElements(123, "foo", "foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotElements(123, "foo")
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestArrayContains(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, "foo"})

	value.Contains(123)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContains(123)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Contains("foo", 123)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContains("foo", 123)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Contains("foo", "foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContains("foo", "foo")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Contains(123, "foo", "FOO")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContains(123, "foo", "FOO")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContains("FOO")
	value.chain.assertOK(t)
	value.chain.reset()

	value.Contains([]interface{}{123, "foo"})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContains([]interface{}{123, "foo"})
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestArrayContainsOnly(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("without_duplicates", func(t *testing.T) {
		value := NewArray(reporter, []interface{}{123, "foo"})

		value.ContainsOnly(123)
		value.chain.assertFailed(t)
		value.chain.reset()

		value.NotContainsOnly(123)
		value.chain.assertOK(t)
		value.chain.reset()

		value.ContainsOnly("foo")
		value.chain.assertFailed(t)
		value.chain.reset()

		value.NotContainsOnly("foo")
		value.chain.assertOK(t)
		value.chain.reset()

		value.ContainsOnly(123, "foo", "foo")
		value.chain.assertOK(t)
		value.chain.reset()

		value.NotContainsOnly(123, "foo", "foo")
		value.chain.assertFailed(t)
		value.chain.reset()

		value.ContainsOnly(123, "foo")
		value.chain.assertOK(t)
		value.chain.reset()

		value.NotContainsOnly(123, "foo")
		value.chain.assertFailed(t)
		value.chain.reset()

		value.ContainsOnly("foo", 123)
		value.chain.assertOK(t)
		value.chain.reset()

		value.NotContainsOnly("foo", 123)
		value.chain.assertFailed(t)
		value.chain.reset()
	})

	t.Run("with_duplicates", func(t *testing.T) {
		value := NewArray(reporter, []interface{}{123, "foo", "foo"})

		value.ContainsOnly(123, "foo")
		value.chain.assertOK(t)
		value.chain.reset()

		value.NotContainsOnly(123, "foo")
		value.chain.assertFailed(t)
		value.chain.reset()

		value.ContainsOnly(123, 123, "foo")
		value.chain.assertOK(t)
		value.chain.reset()

		value.NotContainsOnly(123, 123, "foo")
		value.chain.assertFailed(t)
		value.chain.reset()

		value.ContainsOnly(123, "foo", "foo")
		value.chain.assertOK(t)
		value.chain.reset()

		value.NotContainsOnly(123, "foo", "foo")
		value.chain.assertFailed(t)
		value.chain.reset()

		value.ContainsOnly("foo", 123, "foo")
		value.chain.assertOK(t)
		value.chain.reset()

		value.NotContainsOnly("foo", 123, "foo")
		value.chain.assertFailed(t)
		value.chain.reset()
	})
}

func TestArrayContainsAny(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, "foo"})

	value.ContainsAny(123)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsAny(123)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsAny("foo", 123)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsAny("foo", 123)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsAny("foo", "foo")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsAny("foo", "foo")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsAny(123, "foo", "FOO")
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsAny(123, "foo", "FOO")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsAny("FOO")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsAny("FOO")
	value.chain.assertOK(t)
	value.chain.reset()

	value.ContainsAny([]interface{}{123, "foo"})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsAny([]interface{}{123, "foo"})
	value.chain.assertOK(t)
	value.chain.reset()
}

func TestArrayConvertEqual(t *testing.T) {
	type (
		myArray []interface{}
		myInt   int
	)

	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, 456})

	assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

	value.Equal(myArray{myInt(123), 456.0})
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotEqual(myArray{myInt(123), 456.0})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Equal([]interface{}{"123", "456"})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual([]interface{}{"123", "456"})
	value.chain.assertOK(t)
	value.chain.reset()

	value.Equal(nil)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotEqual(nil)
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestArrayConvertElements(t *testing.T) {
	type (
		myInt int
	)

	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, 456})

	assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

	value.Elements(myInt(123), 456.0)
	value.chain.assertOK(t)
	value.chain.reset()

	value.Elements(func() {})
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestArrayConvertContains(t *testing.T) {
	type (
		myInt int
	)

	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, 456})

	assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

	value.Contains(myInt(123), 456.0)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContains(myInt(123), 456.0)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsOnly(myInt(123), 456.0)
	value.chain.assertOK(t)
	value.chain.reset()

	value.ContainsAny(myInt(123), 456.0)
	value.chain.assertOK(t)
	value.chain.reset()

	value.NotContainsAny(myInt(123), 456.0)
	value.chain.assertFailed(t)
	value.chain.reset()

	value.Contains("123")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContains("123")
	value.chain.assertOK(t)
	value.chain.reset()

	value.ContainsOnly("123.0", "456.0")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsAny("123.0", "456.0")
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsAny("123.0", "456.0")
	value.chain.assertOK(t)
	value.chain.reset()

	value.Contains(func() {})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContains(func() {})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsOnly(func() {})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.ContainsAny(func() {})
	value.chain.assertFailed(t)
	value.chain.reset()

	value.NotContainsAny(func() {})
	value.chain.assertFailed(t)
	value.chain.reset()
}

func TestArrayEvery(t *testing.T) {
	t.Run("Check validation", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{2, 4, 6})
		array.Every(func(_ int, val *Value) {
			if v, ok := val.Raw().(float64); ok {
				newValue(val.chain, int(v)%2).Equal(0)
			}
		})
	})

	t.Run("Empty array", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{})
		array.Every(func(_ int, val *Value) {})
	})

	t.Run("Test correct index", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{1, 2, 3})
		array.Every(
			func(idx int, val *Value) {
				if v, ok := val.Raw().(float64); ok {
					if int(v) != idx+1 {
						val.chain.fail(AssertionFailure{
							Type: AssertEqual,
							Errors: []error{
								errors.New("invalid index"),
							},
						})
					}
				}
			},
		)
	})

	t.Run("Assertion failed for any", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{"foo", "", "bar"})
		invoked := 0
		array.Every(func(_ int, val *Value) {
			invoked++
			val.String().NotEmpty()
		})
		assert.Equal(t, true, array.chain.failed())
		assert.Equal(t, 3, invoked)
	})

	t.Run("Assertion failed for all", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{"", "", ""})
		invoked := 0
		array.Every(func(_ int, val *Value) {
			invoked++
			val.String().NotEmpty()
		})
		assert.Equal(t, true, array.chain.failed())
		assert.Equal(t, 3, invoked)
	})
}
