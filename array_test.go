package httpexpect

import (
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
		value.Filter(func(_ int, val *Value) bool {
			val.String().NotEmpty()
			return true
		})
		value.Transform(func(index int, value interface{}) interface{} {
			return nil
		})
	}

	t.Run("failed_chain", func(t *testing.T) {
		chain := newMockChain(t)
		chain.fail(mockFailure())

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
		chain.fail(mockFailure())

		value := newArray(chain, nil)

		check(value)
	})
}

func TestArrayConstructors(t *testing.T) {
	testValue := []interface{}{"Foo", 123}

	t.Run("Constructor without config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewArray(reporter, testValue)
		value.Equal(testValue)
		value.chain.assertNotFailed(t)
	})

	t.Run("Constructor with config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewArrayC(Config{
			Reporter: reporter,
		}, testValue)
		value.Equal(testValue)
		value.chain.assertNotFailed(t)
	})
}

func TestArrayGetters(t *testing.T) {
	reporter := newMockReporter(t)

	a := []interface{}{"foo", 123.0}

	value := NewArray(reporter, a)

	assert.Equal(t, a, value.Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, a, value.Path("$").Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Schema(`{"type": "array"}`)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Schema(`{"type": "object"}`)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, 2.0, value.Length().Raw())

	assert.Equal(t, "foo", value.Element(0).Raw())
	assert.Equal(t, 123.0, value.Element(1).Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, nil, value.Element(2).Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	it := value.Iter()
	assert.Equal(t, 2, len(it))
	assert.Equal(t, "foo", it[0].Raw())
	assert.Equal(t, 123.0, it[1].Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, "foo", value.First().Raw())
	assert.Equal(t, 123.0, value.Last().Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestArrayEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewArray(reporter, nil)

	_ = value1
	value1.chain.assertFailed(t)
	value1.chain.clearFailed()

	value2 := NewArray(reporter, []interface{}{})

	value2.Empty()
	value2.chain.assertNotFailed(t)
	value2.chain.clearFailed()

	value2.NotEmpty()
	value2.chain.assertFailed(t)
	value2.chain.clearFailed()

	value3 := NewArray(reporter, []interface{}{""})

	value3.Empty()
	value3.chain.assertFailed(t)
	value3.chain.clearFailed()

	value3.NotEmpty()
	value3.chain.assertNotFailed(t)
	value3.chain.clearFailed()
}

func TestArrayEmptyGetters(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{})

	assert.NotNil(t, value.Element(0))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.NotNil(t, value.First())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.NotNil(t, value.Last())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	assert.NotNil(t, value.Iter())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestArrayEqualEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{})

	assert.Equal(t, []interface{}{}, value.Raw())

	value.Equal([]interface{}{})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Equal([]interface{}{""})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{""})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestArrayEqualNotEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{"foo", "bar"})

	assert.Equal(t, []interface{}{"foo", "bar"}, value.Raw())

	value.Equal([]interface{}{})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Equal([]interface{}{"foo"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{"foo"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Equal([]interface{}{"bar", "foo"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{"bar", "foo"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Equal([]interface{}{"foo", "bar"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{"foo", "bar"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()
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
	value1.chain.assertNotFailed(t)
	value1.chain.clearFailed()

	value1.Equal([]string{"bar", "foo"})
	value1.chain.assertFailed(t)
	value1.chain.clearFailed()

	value1.NotEqual([]string{"foo", "bar"})
	value1.chain.assertFailed(t)
	value1.chain.clearFailed()

	value1.NotEqual([]string{"bar", "foo"})
	value1.chain.assertNotFailed(t)
	value1.chain.clearFailed()

	value2.Equal([]int{123, 456})
	value2.chain.assertNotFailed(t)
	value2.chain.clearFailed()

	value2.Equal([]int{456, 123})
	value2.chain.assertFailed(t)
	value2.chain.clearFailed()

	value2.NotEqual([]int{123, 456})
	value2.chain.assertFailed(t)
	value2.chain.clearFailed()

	value2.NotEqual([]int{456, 123})
	value2.chain.assertNotFailed(t)
	value2.chain.clearFailed()

	type S struct {
		Foo int `json:"foo"`
	}

	value3.Equal([]S{{123}, {456}})
	value3.chain.assertNotFailed(t)
	value3.chain.clearFailed()

	value3.Equal([]S{{456}, {123}})
	value3.chain.assertFailed(t)
	value3.chain.clearFailed()

	value3.NotEqual([]S{{123}, {456}})
	value3.chain.assertFailed(t)
	value3.chain.clearFailed()

	value3.NotEqual([]S{{456}, {123}})
	value3.chain.assertNotFailed(t)
	value3.chain.clearFailed()
}

func TestArrayEqualUnordered(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("without_duplicates", func(t *testing.T) {
		value := NewArray(reporter, []interface{}{123, "foo"})

		value.EqualUnordered([]interface{}{123})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{123})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.EqualUnordered([]interface{}{"foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{"foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.EqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.EqualUnordered([]interface{}{123, "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{123, "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.EqualUnordered([]interface{}{"foo", 123})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{"foo", 123})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("with_duplicates", func(t *testing.T) {
		value := NewArray(reporter, []interface{}{123, "foo", "foo"})

		value.EqualUnordered([]interface{}{123, "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{123, "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.EqualUnordered([]interface{}{123, 123, "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{123, 123, "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.EqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.EqualUnordered([]interface{}{"foo", 123, "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{"foo", 123, "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestArrayElements(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, "foo"})

	value.Elements(123)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Elements("foo")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Elements("foo", 123)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Elements(123, "foo", "foo")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Elements(123, "foo")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestArrayNotElements(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, "foo"})

	value.NotElements(123)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotElements("foo")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotElements("foo", 123)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotElements(123, "foo", "foo")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotElements(123, "foo")
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestArrayContains(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, "foo"})

	value.Contains(123)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContains(123)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Contains("foo", 123)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContains("foo", 123)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Contains("foo", "foo")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContains("foo", "foo")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Contains(123, "foo", "FOO")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContains(123, "foo", "FOO")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContains("FOO")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Contains([]interface{}{123, "foo"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContains([]interface{}{123, "foo"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestArrayContainsOnly(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("without_duplicates", func(t *testing.T) {
		value := NewArray(reporter, []interface{}{123, "foo"})

		value.ContainsOnly(123)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotContainsOnly(123)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.ContainsOnly("foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotContainsOnly("foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.ContainsOnly(123, "foo", "foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsOnly(123, "foo", "foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ContainsOnly(123, "foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsOnly(123, "foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ContainsOnly("foo", 123)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsOnly("foo", 123)
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("with_duplicates", func(t *testing.T) {
		value := NewArray(reporter, []interface{}{123, "foo", "foo"})

		value.ContainsOnly(123, "foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsOnly(123, "foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ContainsOnly(123, 123, "foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsOnly(123, 123, "foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ContainsOnly(123, "foo", "foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsOnly(123, "foo", "foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ContainsOnly("foo", 123, "foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsOnly("foo", 123, "foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestArrayContainsAny(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, "foo"})

	value.ContainsAny(123)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsAny(123)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsAny("foo", 123)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsAny("foo", 123)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsAny("foo", "foo")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsAny("foo", "foo")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsAny(123, "foo", "FOO")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsAny(123, "foo", "FOO")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsAny("FOO")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContainsAny("FOO")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.ContainsAny([]interface{}{123, "foo"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContainsAny([]interface{}{123, "foo"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
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
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(myArray{myInt(123), 456.0})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Equal([]interface{}{"123", "456"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{"123", "456"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Equal(nil)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(nil)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestArrayConvertElements(t *testing.T) {
	type (
		myInt int
	)

	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, 456})

	assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

	value.Elements(myInt(123), 456.0)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Elements(func() {})
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestArrayConvertContains(t *testing.T) {
	type (
		myInt int
	)

	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, 456})

	assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

	value.Contains(myInt(123), 456.0)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContains(myInt(123), 456.0)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsOnly(myInt(123), 456.0)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.ContainsAny(myInt(123), 456.0)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotContainsAny(myInt(123), 456.0)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Contains("123")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContains("123")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.ContainsOnly("123.0", "456.0")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsAny("123.0", "456.0")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContainsAny("123.0", "456.0")
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Contains(func() {})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContains(func() {})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsOnly(func() {})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.ContainsAny(func() {})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotContainsAny(func() {})
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestArrayEvery(t *testing.T) {
	t.Run("Check validation", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{2, 4, 6})
		array.Every(func(_ int, val *Value) {
			if v, ok := val.Raw().(float64); ok {
				assert.Equal(ts, 0, int(v)%2)
			}
		})
		array.chain.assertNotFailed(ts)
	})

	t.Run("Empty array", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{})
		array.Every(func(_ int, val *Value) {})
		array.chain.assertNotFailed(ts)
	})

	t.Run("Test correct index", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{1, 2, 3})
		array.Every(
			func(idx int, val *Value) {
				if v, ok := val.Raw().(float64); ok {
					assert.Equal(ts, idx, int(v)-1)
				}
			},
		)
		array.chain.assertNotFailed(ts)
	})

	t.Run("Assertion failed for any", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{"foo", "", "bar"})
		invoked := 0
		array.Every(func(_ int, val *Value) {
			invoked++
			val.String().NotEmpty()
		})
		assert.Equal(t, 3, invoked)
		array.chain.assertFailed(ts)
	})

	t.Run("Assertion failed for all", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{"", "", ""})
		invoked := 0
		array.Every(func(_ int, val *Value) {
			invoked++
			val.String().NotEmpty()
		})
		assert.Equal(t, 3, invoked)
		array.chain.assertFailed(ts)
	})
}

func TestArrayTransform(t *testing.T) {
	t.Run("Square Integers", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{2, 4, 6})
		newArray := array.Transform(func(_ int, val interface{}) interface{} {
			if v, ok := val.(float64); ok {
				return int(v) * int(v)
			}
			ts.Errorf("failed transformation")
			return nil
		})
		assert.Equal(t, []interface{}{float64(4), float64(16), float64(36)}, newArray.value)
		newArray.chain.assertNotFailed(ts)
	})

	t.Run("Chain fail on nil function value", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{2, 4, 6})
		newArray := array.Transform(nil)
		newArray.chain.assertFailed(reporter)
	})

	t.Run("Empty array", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{})
		newArray := array.Transform(func(_ int, _ interface{}) interface{} {
			ts.Errorf("failed transformation")
			return nil
		})
		newArray.chain.assertNotFailed(ts)
	})

	t.Run("Test correct index", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{1, 2, 3})
		newArray := array.Transform(
			func(idx int, val interface{}) interface{} {
				if v, ok := val.(float64); ok {
					assert.Equal(ts, idx, int(v)-1)
				}
				return val
			},
		)
		assert.Equal(t, []interface{}{float64(1), float64(2), float64(3)}, newArray.value)
		newArray.chain.assertNotFailed(ts)
	})
}

func TestArrayFilter(t *testing.T) {
	t.Run("Filter an array of elements of the same type and validate", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, 3, 4, 5, 6})
		filteredArray := array.Filter(func(index int, value *Value) bool {
			return value.Raw() != 2.0 && value.Raw() != 5.0
		})
		assert.Equal(t, []interface{}{1.0, 3.0, 4.0, 6.0}, filteredArray.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, 2.0, 3.0, 4.0, 5.0, 6.0})

		array.chain.assertNotFailed(t)
		filteredArray.chain.assertNotFailed(t)
	})

	t.Run("Filter throws when an assertion within predicate fails", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1.0, "foo", "bar", 4.0, "baz", 6.0})
		filteredArray := array.Filter(func(index int, value *Value) bool {
			stringifiedValue := value.String().NotEmpty().Raw()
			return stringifiedValue != "bar"
		})
		assert.Equal(t, []interface{}{"foo", "baz"}, filteredArray.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, "foo", "bar", 4.0, "baz", 6.0})

		array.chain.assertNotFailed(t)
		filteredArray.chain.assertNotFailed(t)
	})

	t.Run("Filter an array of different types and validate", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{"foo", "bar", true, 1.0})
		filteredArray := array.Filter(func(index int, value *Value) bool {
			return value.Raw() != "bar"
		})
		assert.Equal(t, []interface{}{"foo", true, 1.0}, filteredArray.Raw())
		assert.Equal(t, array.Raw(), []interface{}{"foo", "bar", true, 1.0})

		array.chain.assertNotFailed(t)
		filteredArray.chain.assertNotFailed(t)
	})

	t.Run("Filter an empty array", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})
		filteredArray := array.Filter(func(index int, value *Value) bool {
			return false
		})
		assert.Equal(t, []interface{}{}, filteredArray.Raw())
		assert.Equal(t, array.Raw(), []interface{}{})

		array.chain.assertNotFailed(t)
		filteredArray.chain.assertNotFailed(t)
	})

	t.Run("Filter returns an empty non-nil array if no items are passed",
		func(ts *testing.T) {
			reporter := newMockReporter(t)
			array := NewArray(reporter, []interface{}{"foo", "bar", true, 1.0})
			filteredArray := array.Filter(func(index int, value *Value) bool {
				return false
			})
			assert.Equal(t, []interface{}{}, filteredArray.Raw())
			assert.Equal(t, array.Raw(), []interface{}{"foo", "bar", true, 1.0})

			array.chain.assertNotFailed(t)
			filteredArray.chain.assertNotFailed(t)
		})
}
