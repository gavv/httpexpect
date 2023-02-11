package httpexpect

import (
	"encoding/json"
	"math/big"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArray_FailedChain(t *testing.T) {
	check := func(value *Array) {
		value.chain.assertFailed(t)

		value.Path("$").chain.assertFailed(t)
		value.Schema("")
		value.Alias("foo")

		var target interface{}
		value.Decode(&target)

		value.Length().chain.assertFailed(t)
		value.Element(0).chain.assertFailed(t)
		value.First().chain.assertFailed(t)
		value.Last().chain.assertFailed(t)

		value.IsEmpty()
		value.NotEmpty()
		value.IsEqual([]interface{}{})
		value.NotEqual([]interface{}{})
		value.IsEqualUnordered([]interface{}{})
		value.NotEqualUnordered([]interface{}{})
		value.InList([]interface{}{})
		value.NotInList([]interface{}{})
		value.ConsistsOf("foo")
		value.NotConsistsOf("foo")
		value.Contains("foo")
		value.NotContains("foo")
		value.ContainsAll("foo")
		value.NotContainsAll("foo")
		value.ContainsAny("foo")
		value.NotContainsAny("foo")
		value.ContainsOnly("foo")
		value.NotContainsOnly("foo")

		assert.NotNil(t, value.Iter())
		assert.Equal(t, 0, len(value.Iter()))

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
		value.Find(func(index int, value *Value) bool {
			value.String().NotEmpty()
			return true
		})
		value.FindAll(func(index int, value *Value) bool {
			value.String().NotEmpty()
			return true
		})
		value.NotFind(func(index int, value *Value) bool {
			value.String().NotEmpty()
			return true
		})
	}

	t.Run("failed_chain", func(t *testing.T) {
		chain := newMockChain(t)
		chain.setFailed()

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
		chain.setFailed()

		value := newArray(chain, nil)

		check(value)
	})
}

func TestArray_Constructors(t *testing.T) {
	testValue := []interface{}{"Foo", 123}

	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewArray(reporter, testValue)
		value.IsEqual(testValue)
		value.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewArrayC(Config{
			Reporter: reporter,
		}, testValue)
		value.IsEqual(testValue)
		value.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newArray(chain, testValue)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestArray_Decode(t *testing.T) {
	t.Run("Decode into slice of empty interfaces", func(t *testing.T) {
		reporter := newMockReporter(t)

		testValue := []interface{}{"Foo", json.Number("123.0")}
		arr := NewArray(reporter, testValue)

		var target []interface{}
		arr.Decode(&target)

		arr.chain.assertNotFailed(t)
		assert.Equal(t, testValue, target)
	})

	t.Run("Decode into slice of structs", func(t *testing.T) {
		reporter := newMockReporter(t)

		type S struct {
			Foo int `json:"foo"`
		}

		actualStruct := []S{{123}, {456}}
		testValue := []interface{}{
			map[string]interface{}{
				"foo": json.Number("123"),
			},
			map[string]interface{}{
				"foo": json.Number("456"),
			},
		}
		arr := NewArray(reporter, testValue)

		var target []S
		arr.Decode(&target)

		arr.chain.assertNotFailed(t)
		assert.Equal(t, actualStruct, target)
	})

	t.Run("Decode into empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		testValue := []interface{}{"Foo", json.Number("123.0")}
		arr := NewArray(reporter, testValue)

		var target interface{}
		arr.Decode(&target)

		arr.chain.assertNotFailed(t)
		assert.Equal(t, testValue, target)
	})

	t.Run("Target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		testValue := []interface{}{"Foo", 123}
		arr := NewArray(reporter, testValue)

		arr.Decode(456)
		arr.chain.assertFailed(t)
	})

	t.Run("Target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		testValue := []interface{}{"Foo", 123.0}
		arr := NewArray(reporter, testValue)

		arr.Decode(nil)

		arr.chain.assertFailed(t)
	})
}

func TestArray_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewArray(reporter, []interface{}{1, 2})
	assert.Equal(t, []string{"Array()"}, value1.chain.context.Path)
	assert.Equal(t, []string{"Array()"}, value1.chain.context.AliasedPath)

	value2 := value1.Alias("foo")
	assert.Equal(t, []string{"Array()"}, value2.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value2.chain.context.AliasedPath)

	value3 := value2.Filter(func(index int, value *Value) bool {
		return value.Number().Raw() > 1
	})
	assert.Equal(t, []string{"Array()", "Filter()"}, value3.chain.context.Path)
	assert.Equal(t, []string{"foo", "Filter()"}, value3.chain.context.AliasedPath)
}

func TestArray_Getters(t *testing.T) {
	reporter := newMockReporter(t)

	a := []interface{}{"foo", json.Number("123.0")}

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
	assert.Equal(t, json.Number("123.0"), value.Element(1).Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, nil, value.Element(2).Raw())
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	it := value.Iter()
	assert.Equal(t, 2, len(it))
	assert.Equal(t, "foo", it[0].Raw())
	assert.Equal(t, json.Number("123.0"), it[1].Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, "foo", value.First().Raw())
	assert.Equal(t, json.Number("123.0"), value.Last().Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestArray_Empty(t *testing.T) {
	reporter := newMockReporter(t)

	value1 := NewArray(reporter, nil)

	_ = value1
	value1.chain.assertFailed(t)
	value1.chain.clearFailed()

	value2 := NewArray(reporter, []interface{}{})

	value2.IsEmpty()
	value2.chain.assertNotFailed(t)
	value2.chain.clearFailed()

	value2.NotEmpty()
	value2.chain.assertFailed(t)
	value2.chain.clearFailed()

	value3 := NewArray(reporter, []interface{}{""})

	value3.IsEmpty()
	value3.chain.assertFailed(t)
	value3.chain.clearFailed()

	value3.NotEmpty()
	value3.chain.assertNotFailed(t)
	value3.chain.clearFailed()
}

func TestArray_EmptyGetters(t *testing.T) {
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

func TestArray_EqualEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{})

	assert.Equal(t, []interface{}{}, value.Raw())

	value.IsEqual([]interface{}{})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.IsEqual([]interface{}{""})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{""})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestArray_EqualNotEmpty(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{"foo", "bar"})

	assert.Equal(t, []interface{}{"foo", "bar"}, value.Raw())

	value.IsEqual([]interface{}{})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsEqual([]interface{}{"foo"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{"foo"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsEqual([]interface{}{"bar", "foo"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{"bar", "foo"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsEqual([]interface{}{"foo", "bar"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual([]interface{}{"foo", "bar"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestArray_EqualTypes(t *testing.T) {
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

	value1.IsEqual([]string{"foo", "bar"})
	value1.chain.assertNotFailed(t)
	value1.chain.clearFailed()

	value1.IsEqual([]string{"bar", "foo"})
	value1.chain.assertFailed(t)
	value1.chain.clearFailed()

	value1.NotEqual([]string{"foo", "bar"})
	value1.chain.assertFailed(t)
	value1.chain.clearFailed()

	value1.NotEqual([]string{"bar", "foo"})
	value1.chain.assertNotFailed(t)
	value1.chain.clearFailed()

	value2.IsEqual([]int{123, 456})
	value2.chain.assertNotFailed(t)
	value2.chain.clearFailed()

	value2.IsEqual([]int{456, 123})
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

	value3.IsEqual([]S{{123}, {456}})
	value3.chain.assertNotFailed(t)
	value3.chain.clearFailed()

	value3.IsEqual([]S{{456}, {123}})
	value3.chain.assertFailed(t)
	value3.chain.clearFailed()

	value3.NotEqual([]S{{123}, {456}})
	value3.chain.assertFailed(t)
	value3.chain.clearFailed()

	value3.NotEqual([]S{{456}, {123}})
	value3.chain.assertNotFailed(t)
	value3.chain.clearFailed()
}

func TestArray_EqualUnordered(t *testing.T) {
	reporter := newMockReporter(t)

	t.Run("without_duplicates", func(t *testing.T) {
		value := NewArray(reporter, []interface{}{123, "foo"})

		value.IsEqualUnordered([]interface{}{123})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{123})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.IsEqualUnordered([]interface{}{"foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{"foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.IsEqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.IsEqualUnordered([]interface{}{123, "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{123, "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsEqualUnordered([]interface{}{"foo", 123})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{"foo", 123})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("with_duplicates", func(t *testing.T) {
		value := NewArray(reporter, []interface{}{123, "foo", "foo"})

		value.IsEqualUnordered([]interface{}{123, "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{123, "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.IsEqualUnordered([]interface{}{123, 123, "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{123, 123, "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.IsEqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsEqualUnordered([]interface{}{"foo", 123, "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered([]interface{}{"foo", 123, "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestArray_InList(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{"foo", "bar"})

	assert.Equal(t, []interface{}{"foo", "bar"}, value.Raw())

	value.InList()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList("foo", "bar")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList("foo", "bar")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList([]interface{}{})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList([]interface{}{})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InList([]interface{}{"bar", "foo"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList([]interface{}{"bar", "foo"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InList([]interface{}{"bar", "foo"}, []interface{}{"foo", "bar"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInList([]interface{}{"bar", "foo"}, []interface{}{"foo", "bar"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList([]interface{}{"bar", "foo"}, []interface{}{"FOO", "BAR"})
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList([]interface{}{"bar", "foo"}, []interface{}{"FOO", "BAR"})
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InList([]interface{}{"bar", "foo"}, "NOT ARRAY")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList([]interface{}{"bar", "foo"}, "NOT ARRAY")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList([]interface{}{"foo", "bar"}, "NOT ARRAY")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList([]interface{}{"foo", "bar"}, "NOT ARRAY")
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestArray_ConsistsOf(t *testing.T) {
	t.Run("ConsistsOf", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, "foo"})

		value.ConsistsOf(123)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ConsistsOf("foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ConsistsOf("foo", 123)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ConsistsOf(123, "foo", "foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ConsistsOf(123, "foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()
	})

	t.Run("NotConsistsOf", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, "foo"})

		value.NotConsistsOf(123)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotConsistsOf("foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotConsistsOf("foo", 123)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotConsistsOf(123, "foo", "foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotConsistsOf(123, "foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestArray_Contains(t *testing.T) {
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

func TestArray_ContainsAny(t *testing.T) {
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

func TestArray_ContainsOnly(t *testing.T) {
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

func TestArray_ConvertEqual(t *testing.T) {
	type (
		myArray []interface{}
		myInt   json.Number
	)

	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, 456})

	assert.Equal(t, []interface{}{json.Number("123"), json.Number("456")}, value.Raw())

	// value.IsEqual(myArray{myInt("123"), json.Number("456")})
	// value.chain.assertNotFailed(t)
	// value.chain.clearFailed()

	// value.NotEqual(myArray{myInt("123"), json.Number("456.0")})
	// value.chain.assertFailed(t)
	// value.chain.clearFailed()
	//
	// value.IsEqual([]interface{}{"123", "456.0"})
	// value.chain.assertFailed(t)
	// value.chain.clearFailed()
	//
	// value.NotEqual([]interface{}{"123", "456.0"})
	// value.chain.assertNotFailed(t)
	// value.chain.clearFailed()
	//
	// value.IsEqual(nil)
	// value.chain.assertFailed(t)
	// value.chain.clearFailed()
	//
	// value.NotEqual(nil)
	// value.chain.assertFailed(t)
	// value.chain.clearFailed()
}

func TestArray_ConvertConsistsOf(t *testing.T) {
	type (
		myInt int
	)

	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, 456})

	assert.Equal(t, []interface{}{json.Number("123"), json.Number("456")}, value.Raw())

	value.ConsistsOf(myInt(123), 456.0)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.ConsistsOf(func() {})
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestArray_ConvertContains(t *testing.T) {
	type (
		myInt int
	)

	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{123, 456})

	assert.Equal(t, []interface{}{json.Number("123"), json.Number("456")}, value.Raw())

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

func TestArray_Every(t *testing.T) {
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

func TestArray_Transform(t *testing.T) {
	t.Run("Square Integers", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{
			big.NewFloat(0).SetInt64(2),
			big.NewFloat(0).SetInt64(4),
			big.NewFloat(0).SetInt64(6),
		})
		newArray := array.Transform(func(_ int, val interface{}) interface{} {
			if v, ok := val.(string); ok {
				value, _ := big.NewFloat(0).SetString(v)
				return big.NewFloat(0).Mul(value, value)
			}
			ts.Errorf("failed transformation %T", val)
			return big.NewFloat(0)
		})
		assert.Equal(t, []interface{}{
			"4",
			"16",
			"36",
		}, newArray.value)
		newArray.chain.assertNotFailed(ts)
	})

	t.Run("Chain fail on nil function value", func(ts *testing.T) {
		reporter := newMockReporter(ts)
		array := NewArray(reporter, []interface{}{2, 4, 6})
		newArray := array.Transform(nil)
		newArray.chain.assertFailed(t)
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
		assert.Equal(t, []interface{}{
			json.Number("1"),
			json.Number("2"),
			json.Number("3"),
		}, newArray.value)
		newArray.chain.assertNotFailed(ts)
	})
}

func TestArray_Filter(t *testing.T) {
	t.Run("Filter an array of elements of the same type and validate", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, 3, 4, 5, 6})
		filteredArray := array.Filter(func(index int, value *Value) bool {
			return value.Raw() != json.Number("2") && value.Raw() != json.Number("5")
		})
		assert.Equal(t, []interface{}{
			json.Number("1"),
			json.Number("3"),
			json.Number("4"),
			json.Number("6"),
		}, filteredArray.Raw())
		assert.Equal(t, array.Raw(), []interface{}{
			json.Number("1"),
			json.Number("2"),
			json.Number("3"),
			json.Number("4"),
			json.Number("5"),
			json.Number("6"),
		})

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
		assert.Equal(t, []interface{}{
			json.Number("1"),
			"foo", "bar",
			json.Number("4"),
			"baz",
			json.Number("6"),
		}, array.Raw())

		array.chain.assertNotFailed(t)
		filteredArray.chain.assertNotFailed(t)
	})

	t.Run("Filter an array of different types and validate", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{"foo", "bar", true, 1.0})
		filteredArray := array.Filter(func(index int, value *Value) bool {
			return value.Raw() != "bar"
		})
		assert.Equal(t, []interface{}{"foo", true, json.Number("1")}, filteredArray.Raw())
		assert.Equal(t, []interface{}{"foo", "bar", true, json.Number("1")}, array.Raw())

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
			assert.Equal(t, []interface{}{"foo", "bar", true, json.Number("1")}, array.Raw())

			array.chain.assertNotFailed(t)
			filteredArray.chain.assertNotFailed(t)
		})
}

func TestArray_Find(t *testing.T) {
	t.Run("Elements of the same type", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, 3, 4, 5, 6})
		foundValue := array.Find(func(index int, value *Value) bool {
			return value.Raw() == json.Number("2")
		})
		assert.Equal(t, json.Number("2"), foundValue.Raw())
		assert.Equal(t, []interface{}{
			json.Number("1"),
			json.Number("2"),
			json.Number("3"),
			json.Number("4"),
			json.Number("5"),
			json.Number("6"),
		}, array.Raw())

		array.chain.assertNotFailed(t)
		foundValue.chain.assertNotFailed(t)
	})

	t.Run("Elements of multiple types", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, "foo", true, "bar"})
		foundValue := array.Find(func(index int, value *Value) bool {
			stringifiedValue := value.String().NotEmpty().Raw()
			return stringifiedValue == "bar"
		})
		assert.Equal(t, "bar", foundValue.Raw())
		assert.Equal(t, array.Raw(), []interface{}{json.Number("1"), "foo", true, "bar"})

		array.chain.assertNotFailed(t)
		foundValue.chain.assertNotFailed(t)
	})

	t.Run("First match", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, "foo", true, "bar"})
		foundValue := array.Find(func(index int, value *Value) bool {
			stringifiedValue := value.String().Raw()
			return stringifiedValue != ""
		})
		assert.Equal(t, "foo", foundValue.Raw())
		assert.Equal(t, []interface{}{json.Number("1"), "foo", true, "bar"}, array.Raw())

		array.chain.assertNotFailed(t)
		foundValue.chain.assertNotFailed(t)
	})

	t.Run("No match", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, "foo", true, "bar"})
		foundValue := array.Find(func(index int, value *Value) bool {
			return value.Raw() == json.Number("2")
		})
		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, []interface{}{json.Number("1"), "foo", true, "bar"}, array.Raw())

		array.chain.assertFailed(t)
		foundValue.chain.assertFailed(t)
	})

	t.Run("Empty array", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})
		foundValue := array.Find(func(index int, value *Value) bool {
			return value.Raw() == json.Number("2")
		})
		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, []interface{}{}, array.Raw())

		array.chain.assertFailed(t)
		foundValue.chain.assertFailed(t)
	})

	t.Run("Predicate returns true, assertion fails, no match",
		func(ts *testing.T) {
			reporter := newMockReporter(t)
			array := NewArray(reporter, []interface{}{1, 2})
			foundValue := array.Find(func(index int, value *Value) bool {
				value.String()
				return true
			})
			assert.Equal(t, nil, foundValue.Raw())
			assert.Equal(t, []interface{}{json.Number("1"), json.Number("2")}, array.Raw())

			array.chain.assertFailed(t)
			foundValue.chain.assertFailed(t)
		})

	t.Run("Predicate returns true, assertion fails, have match",
		func(ts *testing.T) {
			reporter := newMockReporter(t)
			array := NewArray(reporter, []interface{}{1, 2, "str"})
			foundValue := array.Find(func(index int, value *Value) bool {
				value.String()
				return true
			})
			assert.Equal(t, "str", foundValue.Raw())
			assert.Equal(t, []interface{}{json.Number("1"), json.Number("2"), "str"}, array.Raw())

			array.chain.assertNotFailed(t)
			foundValue.chain.assertNotFailed(t)
		})

	t.Run("Predicate func is nil", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2})
		foundValue := array.Find(nil)
		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, []interface{}{json.Number("1"), json.Number("2")}, array.Raw())

		array.chain.assertFailed(t)
		foundValue.chain.assertFailed(t)
	})
}

func TestArray_FindAll(t *testing.T) {
	t.Run("Elements of the same type", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, 3, 4, 5, 6})
		foundValues := array.FindAll(func(index int, value *Value) bool {
			return value.Raw() == json.Number("2") || value.Raw() == json.Number("5")
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}

		assert.Equal(t, []interface{}{json.Number("2"), json.Number("5")}, actual)
		assert.Equal(t, []interface{}{
			json.Number("1"),
			json.Number("2"),
			json.Number("3"),
			json.Number("4"),
			json.Number("5"),
			json.Number("6"),
		}, array.Raw())

		array.chain.assertNotFailed(t)
		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
		}
	})

	t.Run("Elements of multiple types", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1.0, "foo", true, "bar"})
		foundValues := array.FindAll(func(index int, value *Value) bool {
			stringifiedValue := value.String().Raw()
			return stringifiedValue != ""
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}
		assert.Equal(t, []interface{}{"foo", "bar"}, actual)
		assert.Equal(t, []interface{}{json.Number("1"), "foo", true, "bar"}, array.Raw())

		array.chain.assertNotFailed(t)
		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
		}
	})

	t.Run("No match", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1.0, "foo", true, "bar"})
		foundValues := array.FindAll(func(index int, value *Value) bool {
			return value.Raw() == json.Number("2")
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}
		assert.Equal(t, []interface{}{}, actual)
		assert.Equal(t, []interface{}{json.Number("1"), "foo", true, "bar"}, array.Raw())

		array.chain.assertNotFailed(t)
		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
		}
	})

	t.Run("Empty array", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})
		foundValues := array.FindAll(func(index int, value *Value) bool {
			return value.Raw() == json.Number("2")
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}
		assert.Equal(t, []interface{}{}, actual)
		assert.Equal(t, []interface{}{}, array.Raw())

		array.chain.assertNotFailed(t)
		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
		}
	})

	t.Run("Predicate returns true, assertion fails, no match",
		func(ts *testing.T) {
			reporter := newMockReporter(t)
			array := NewArray(reporter, []interface{}{1, 2})
			foundValues := array.FindAll(func(index int, value *Value) bool {
				value.String()
				return true
			})

			actual := []interface{}{}
			for _, value := range foundValues {
				actual = append(actual, value.Raw())
			}
			assert.Equal(t, []interface{}{}, actual)
			assert.Equal(t, []interface{}{json.Number("1"), json.Number("2")}, array.Raw())

			array.chain.assertNotFailed(t)
			for _, value := range foundValues {
				value.chain.assertNotFailed(t)
			}
		})

	t.Run("Predicate returns true, assertion fails, have matches", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{"foo", 1, 2, "bar"})
		foundValues := array.FindAll(func(index int, value *Value) bool {
			value.String()
			return true
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}
		assert.Equal(t, []interface{}{"foo", "bar"}, actual)
		assert.Equal(t, []interface{}{"foo", json.Number("1"), json.Number("2"), "bar"}, array.Raw())

		array.chain.assertNotFailed(t)
		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
		}
	})

	t.Run("Predicate func is nil", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2})
		foundValues := array.FindAll(nil)

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}
		assert.Equal(t, []interface{}{}, actual)
		assert.Equal(t, []interface{}{json.Number("1"), json.Number("2")}, array.Raw())

		array.chain.assertFailed(t)
		for _, value := range foundValues {
			value.chain.assertFailed(t)
		}
	})
}

func TestArray_NotFind(t *testing.T) {
	t.Run("Succeeds if no element matches predicate", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, "foo", true, "bar"})
		afterArray := array.NotFind(func(index int, value *Value) bool {
			return value.String().Raw() == "baz"
		})
		assert.Same(t, array, afterArray)
		array.chain.assertNotFailed(t)
	})

	t.Run("Fails if there is a match", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, "foo", true, "bar"})
		afterArray := array.NotFind(func(index int, value *Value) bool {
			return value.String().NotEmpty().Raw() == "bar"
		})
		assert.Same(t, array, afterArray)
		array.chain.assertFailed(t)
	})

	t.Run("Empty array", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})
		afterArray := array.NotFind(func(index int, value *Value) bool {
			return value.Raw() == 2.0
		})
		assert.Same(t, array, afterArray)
		array.chain.assertNotFailed(t)
	})

	t.Run("Predicate returns true, assertion fails, no match",
		func(ts *testing.T) {
			reporter := newMockReporter(t)
			array := NewArray(reporter, []interface{}{1, 2})
			afterArray := array.NotFind(func(index int, value *Value) bool {
				value.String()
				return true
			})
			assert.Same(t, array, afterArray)
			array.chain.assertNotFailed(t)
		})

	t.Run("Predicate returns true, assertion fails, have match",
		func(ts *testing.T) {
			reporter := newMockReporter(t)
			array := NewArray(reporter, []interface{}{1, 2, "str"})
			afterArray := array.NotFind(func(index int, value *Value) bool {
				value.String()
				return true
			})
			assert.Same(t, array, afterArray)
			array.chain.assertFailed(t)
		})

	t.Run("Predicate func is nil", func(ts *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2})
		afterArray := array.NotFind(nil)
		assert.Same(t, array, afterArray)
		array.chain.assertFailed(t)
	})
}

func TestArray_IsOrdered(t *testing.T) {
	type args struct {
		values      []interface{}
		less        []func(x, y *Value) bool
		chainFailed bool
	}
	tests := []struct {
		name   string
		args   args
		wantOK bool
	}{
		{
			name: "array boolean ordered",
			args: args{
				values: []interface{}{false, false, true, true},
			},
			wantOK: true,
		},
		{
			name: "array number ordered",
			args: args{
				values: []interface{}{1, 1, 2, 3},
			},
			wantOK: true,
		},
		{
			name: "array string ordered",
			args: args{
				values: []interface{}{"", "a", "b", "ba"},
			},
			wantOK: true,
		},
		{
			name: "array of nil elements",
			args: args{
				values: []interface{}{nil, nil, nil},
			},
			wantOK: true,
		},
		{
			name: "wrong order",
			args: args{
				values: []interface{}{3, 2, 1},
			},
			wantOK: false,
		},
		{
			name: "user-defined less function",
			args: args{
				values: []interface{}{1, 2, 3},
				less: []func(x, y *Value) bool{
					func(x, y *Value) bool {
						valX := x.Number().Raw()
						valY := y.Number().Raw()
						return valX < valY
					},
				},
			},
			wantOK: true,
		},
		{
			name: "invalid - failed type assertion on less function",
			args: args{
				values: []interface{}{1, 2, 3},
				less: []func(x, y *Value) bool{
					func(x, y *Value) bool {
						x.String()
						y.String()
						return false
					},
				},
			},
			wantOK: false,
		},
		{
			name: "invalid - multiple less functions",
			args: args{
				values: []interface{}{1, 2, 3},
				less: []func(x, y *Value) bool{
					func(x, y *Value) bool {
						return false
					},
					func(x, y *Value) bool {
						return true
					},
				},
			},
			wantOK: false,
		},
		{
			name: "invalid - nil less function",
			args: args{
				values: []interface{}{1, 2, 3},
				less: []func(x, y *Value) bool{
					nil,
				},
			},
			wantOK: false,
		},
		{
			name: "invalid - data type not allowed",
			args: args{
				values: []interface{}{[]int{1, 2}, []int{3, 4}, []int{5, 6}},
				less:   []func(x, y *Value) bool{},
			},
			wantOK: false,
		},
		{
			name: "invalid - multiple data types found",
			args: args{
				values: []interface{}{1, "abc", true},
				less:   []func(x, y *Value) bool{},
			},
			wantOK: false,
		},
		{
			name: "empty array",
			args: args{
				values: []interface{}{},
			},
			wantOK: true,
		},
		{
			name: "one element",
			args: args{
				values: []interface{}{1},
			},
			wantOK: true,
		},
		{
			name: "empty array - custom func",
			args: args{
				values: []interface{}{},
				less: []func(x, y *Value) bool{
					func(x, y *Value) bool {
						panic("test")
					},
				},
			},
			wantOK: true,
		},
		{
			name: "one element - custom func",
			args: args{
				values: []interface{}{1},
				less: []func(x, y *Value) bool{
					func(x, y *Value) bool {
						panic("test")
					},
				},
			},
			wantOK: true,
		},
		{
			name: "chain has failed before",
			args: args{
				chainFailed: true,
			},
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			a := NewArray(reporter, tt.args.values)
			a.IsOrdered(tt.args.less...)
			if tt.wantOK {
				a.chain.assertNotFailed(t)
			} else {
				a.chain.assertFailed(t)
			}
			a.chain.clearFailed()
		})
	}
}

func TestArray_NotOrdered(t *testing.T) {
	type args struct {
		values      []interface{}
		less        []func(x, y *Value) bool
		chainFailed bool
	}
	tests := []struct {
		name   string
		args   args
		wantOK bool
	}{
		{
			name: "array boolean not ordered",
			args: args{
				values: []interface{}{true, true, false, false},
			},
			wantOK: true,
		},
		{
			name: "array number not ordered",
			args: args{
				values: []interface{}{3, 1, 1, 2},
			},
			wantOK: true,
		},
		{
			name: "array string not ordered",
			args: args{
				values: []interface{}{"z", "y", "x", ""},
			},
			wantOK: true,
		},
		{
			name: "array of nil elements",
			args: args{
				values: []interface{}{nil, nil, nil},
			},
			wantOK: false,
		},
		{
			name: "array ordered",
			args: args{
				values: []interface{}{1, 2, 3},
			},
			wantOK: false,
		},
		{
			name: "user-defined less function",
			args: args{
				values: []interface{}{1, 2, 3},
				less: []func(x, y *Value) bool{
					func(x, y *Value) bool {
						valX := x.Number().Raw()
						valY := y.Number().Raw()
						return valX >= valY
					},
				},
			},
			wantOK: true,
		},
		{
			name: "invalid - failed type assertion on less function",
			args: args{
				values: []interface{}{1, 2},
				less: []func(x, y *Value) bool{
					func(x, y *Value) bool {
						x.String()
						y.String()
						return false
					},
				},
			},
			wantOK: false,
		},
		{
			name: "invalid - multiple less functions",
			args: args{
				values: []interface{}{1, 2, 3},
				less: []func(x, y *Value) bool{
					func(x, y *Value) bool {
						return false
					},
					func(x, y *Value) bool {
						return true
					},
				},
			},
			wantOK: false,
		},
		{
			name: "invalid - nil less function",
			args: args{
				values: []interface{}{1, 2, 3},
				less: []func(x, y *Value) bool{
					nil,
				},
			},
			wantOK: false,
		},
		{
			name: "invalid - data type not allowed",
			args: args{
				values: []interface{}{[]int{1, 2}, []int{3, 4}, []int{5, 6}},
				less:   []func(x, y *Value) bool{},
			},
			wantOK: false,
		},
		{
			name: "invalid - multiple data types found",
			args: args{
				values: []interface{}{1, "abc", true},
				less:   []func(x, y *Value) bool{},
			},
			wantOK: false,
		},
		{
			name: "empty array",
			args: args{
				values: []interface{}{},
			},
			wantOK: true,
		},
		{
			name: "one element",
			args: args{
				values: []interface{}{1},
			},
			wantOK: true,
		},
		{
			name: "empty array - custom func",
			args: args{
				values: []interface{}{},
				less: []func(x, y *Value) bool{
					func(x, y *Value) bool {
						panic("test")
					},
				},
			},
			wantOK: true,
		},
		{
			name: "one element - custom func",
			args: args{
				values: []interface{}{1},
				less: []func(x, y *Value) bool{
					func(x, y *Value) bool {
						panic("test")
					},
				},
			},
			wantOK: true,
		},
		{
			name: "chain has failed before",
			args: args{
				chainFailed: true,
			},
			wantOK: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)
			a := NewArray(reporter, tt.args.values)
			a.NotOrdered(tt.args.less...)
			if tt.wantOK {
				a.chain.assertNotFailed(t)
			} else {
				a.chain.assertFailed(t)
			}
			a.chain.clearFailed()
		})
	}
}

func TestArray_ComparatorErrors(t *testing.T) {
	t.Run("nil_slice", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, nil)

		assert.Nil(t, fn)
		chain.assertNotFailed(t)
	})

	t.Run("0_elements", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, []interface{}{})

		assert.Nil(t, fn)
		chain.assertNotFailed(t)
	})

	t.Run("1_element", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, []interface{}{
			"test",
		})

		assert.Nil(t, fn)
		chain.assertNotFailed(t)
	})

	t.Run("2_elements_bad_type", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, []interface{}{
			"test",
			make(chan int), // bad type
		})

		assert.Nil(t, fn)
		chain.assertFailed(t)
	})

	t.Run("2_elements_good_types", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, []interface{}{
			"test",
			"test",
		})

		assert.NotNil(t, fn)
		chain.assertNotFailed(t)
	})
}
