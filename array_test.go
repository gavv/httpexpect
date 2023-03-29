package httpexpect

import (
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestArray_FailedChain(t *testing.T) {
	check := func(value *Array) {
		value.chain.assert(t, failure)

		value.Path("$").chain.assert(t, failure)
		value.Schema("")
		value.Alias("foo")

		var target interface{}
		value.Decode(&target)

		value.Length().chain.assert(t, failure)
		value.Value(0).chain.assert(t, failure)
		value.First().chain.assert(t, failure)
		value.Last().chain.assert(t, failure)

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
		value.HasValue(0, nil)
		value.NotHasValue(0, nil)

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

	t.Run("failed chain", func(t *testing.T) {
		chain := newMockChain(t, flagFailed)
		value := newArray(chain, []interface{}{})

		check(value)
	})

	t.Run("nil value", func(t *testing.T) {
		chain := newMockChain(t)
		value := newArray(chain, nil)

		check(value)
	})

	t.Run("failed chain, nil value", func(t *testing.T) {
		chain := newMockChain(t, flagFailed)
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
		value.chain.assert(t, success)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArrayC(Config{
			Reporter: reporter,
		}, testValue)

		value.IsEqual(testValue)
		value.chain.assert(t, success)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)

		value := newArray(chain, testValue)

		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})

	t.Run("invalid value", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, nil)

		value.chain.assert(t, failure)
	})
}

func TestArray_Decode(t *testing.T) {
	t.Run("target is empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		testValue := []interface{}{"Foo", 123.0}
		arr := NewArray(reporter, testValue)

		var target interface{}
		arr.Decode(&target)

		arr.chain.assert(t, success)
		assert.Equal(t, testValue, target)
	})

	t.Run("target is slice of empty interfaces", func(t *testing.T) {
		reporter := newMockReporter(t)

		testValue := []interface{}{"Foo", 123.0}
		arr := NewArray(reporter, testValue)

		var target []interface{}
		arr.Decode(&target)

		arr.chain.assert(t, success)
		assert.Equal(t, testValue, target)
	})

	t.Run("target is slice of structs", func(t *testing.T) {
		reporter := newMockReporter(t)

		type S struct {
			Foo int `json:"foo"`
		}

		actualStruct := []S{{123}, {456}}
		testValue := []interface{}{
			map[string]interface{}{
				"foo": 123,
			},
			map[string]interface{}{
				"foo": 456,
			},
		}
		arr := NewArray(reporter, testValue)

		var target []S
		arr.Decode(&target)

		arr.chain.assert(t, success)
		assert.Equal(t, actualStruct, target)
	})

	t.Run("target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		testValue := []interface{}{"Foo", 123.0}
		arr := NewArray(reporter, testValue)

		arr.Decode(123)

		arr.chain.assert(t, failure)
	})

	t.Run("target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		testValue := []interface{}{"Foo", 123.0}
		arr := NewArray(reporter, testValue)

		arr.Decode(nil)

		arr.chain.assert(t, failure)
	})
}

func TestArray_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewArray(reporter, []interface{}{1, 2})
	assert.Equal(t, []string{"Array()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Array()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Array()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)

	childValue := value.Filter(func(index int, value *Value) bool {
		return value.Number().Raw() > 1
	})
	assert.Equal(t, []string{"Array()", "Filter()"}, childValue.chain.context.Path)
	assert.Equal(t, []string{"foo", "Filter()"}, childValue.chain.context.AliasedPath)
}

func TestArray_Getters(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		data := []interface{}{}

		value := NewArray(reporter, data)

		assert.Equal(t, data, value.Raw())
		value.chain.assert(t, success)
		value.chain.clear()

		assert.Equal(t, data, value.Path("$").Raw())
		value.chain.assert(t, success)
		value.chain.clear()

		value.Schema(`{"type": "array"}`)
		value.chain.assert(t, success)
		value.chain.clear()

		value.Schema(`{"type": "object"}`)
		value.chain.assert(t, failure)
		value.chain.clear()

		assert.Equal(t, 0.0, value.Length().Raw())
		value.chain.assert(t, success)
		value.chain.clear()

		assert.NotNil(t, value.Value(0))
		value.chain.assert(t, failure)
		value.chain.clear()

		assert.NotNil(t, value.First())
		value.chain.assert(t, failure)
		value.chain.clear()

		assert.NotNil(t, value.Last())
		value.chain.assert(t, failure)
		value.chain.clear()

		assert.NotNil(t, value.Iter())
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("not empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		data := []interface{}{"foo", 123.0}

		value := NewArray(reporter, data)

		assert.Equal(t, data, value.Raw())
		value.chain.assert(t, success)
		value.chain.clear()

		assert.Equal(t, data, value.Path("$").Raw())
		value.chain.assert(t, success)
		value.chain.clear()

		value.Schema(`{"type": "array"}`)
		value.chain.assert(t, success)
		value.chain.clear()

		value.Schema(`{"type": "object"}`)
		value.chain.assert(t, failure)
		value.chain.clear()

		assert.Equal(t, 2.0, value.Length().Raw())
		value.chain.assert(t, success)
		value.chain.clear()

		assert.Equal(t, "foo", value.Value(0).Raw())
		assert.Equal(t, 123.0, value.Value(1).Raw())
		value.chain.assert(t, success)
		value.chain.clear()

		assert.Equal(t, nil, value.Value(2).Raw())
		value.chain.assert(t, failure)
		value.chain.clear()

		assert.Equal(t, "foo", value.First().Raw())
		assert.Equal(t, 123.0, value.Last().Raw())
		value.chain.assert(t, success)
		value.chain.clear()

		it := value.Iter()
		assert.Equal(t, 2, len(it))
		assert.Equal(t, "foo", it[0].Raw())
		assert.Equal(t, 123.0, it[1].Raw())
		value.chain.assert(t, success)
		value.chain.clear()
	})
}

func TestArray_IsEmpty(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.IsEmpty()
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEmpty()
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("one empty element", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{""})

		value.IsEmpty()
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEmpty()
		value.chain.assert(t, success)
		value.chain.clear()
	})
}

func TestArray_IsEqual(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		assert.Equal(t, []interface{}{}, value.Raw())

		value.IsEqual([]interface{}{})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqual([]interface{}{})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.IsEqual([]interface{}{""})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual([]interface{}{""})
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("not empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{"foo", "bar"})

		assert.Equal(t, []interface{}{"foo", "bar"}, value.Raw())

		value.IsEqual([]interface{}{})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual([]interface{}{})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqual([]interface{}{"foo"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual([]interface{}{"foo"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqual([]interface{}{"bar", "foo"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual([]interface{}{"bar", "foo"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqual([]interface{}{"foo", "bar"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqual([]interface{}{"foo", "bar"})
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("strings", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{"foo", "bar"})

		value.IsEqual([]string{"foo", "bar"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqual([]string{"bar", "foo"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual([]string{"foo", "bar"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual([]string{"bar", "foo"})
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("numbers", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456})

		value.IsEqual([]int{123, 456})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqual([]int{456, 123})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual([]int{123, 456})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual([]int{456, 123})
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("struct", func(t *testing.T) {
		reporter := newMockReporter(t)
		type S struct {
			Foo int `json:"foo"`
		}

		value := NewArray(reporter, []interface{}{
			map[string]interface{}{
				"foo": 123,
			},
			map[string]interface{}{
				"foo": 456,
			},
		})

		value.IsEqual([]S{{123}, {456}})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqual([]S{{456}, {123}})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual([]S{{123}, {456}})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual([]S{{456}, {123}})
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myArray []interface{}
			myInt   int
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456})

		assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

		value.IsEqual(myArray{myInt(123), 456.0})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqual(myArray{myInt(123), 456.0})
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.IsEqual(nil)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual(nil)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.IsEqual(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqual(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()
	})
}

func TestArray_IsEqualUnordered(t *testing.T) {
	t.Run("without duplicates", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, "foo"})

		value.IsEqualUnordered([]interface{}{123})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqualUnordered([]interface{}{123})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqualUnordered([]interface{}{"foo"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqualUnordered([]interface{}{"foo"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqualUnordered([]interface{}{123, "foo"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqualUnordered([]interface{}{123, "foo"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.IsEqualUnordered([]interface{}{"foo", 123})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqualUnordered([]interface{}{"foo", 123})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.IsEqualUnordered([]interface{}{"foo", 1234})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqualUnordered([]interface{}{"foo", 1234})
		value.chain.assert(t, success)
		value.chain.clear()

	})

	t.Run("with duplicates", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, "foo", "foo"})

		value.IsEqualUnordered([]interface{}{123, "foo"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqualUnordered([]interface{}{123, "foo"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqualUnordered([]interface{}{123, 123, "foo"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqualUnordered([]interface{}{123, 123, "foo"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.IsEqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqualUnordered([]interface{}{123, "foo", "foo"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.IsEqualUnordered([]interface{}{"foo", 123, "foo"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqualUnordered([]interface{}{"foo", 123, "foo"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.IsEqualUnordered([]interface{}{123})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqualUnordered([]interface{}{123})
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myArray []interface{}
			myInt   int
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456, "foo"})

		assert.Equal(t, []interface{}{123.0, 456.0, "foo"}, value.Raw())

		value.IsEqualUnordered(myArray{myInt(456), 123.0, "foo"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotEqualUnordered(myArray{myInt(456), 123.0, "foo"})
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.IsEqualUnordered(nil)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqualUnordered(nil)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.IsEqualUnordered(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotEqualUnordered(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()
	})
}

func TestArray_InList(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{"foo", "bar"})

		value.InList("foo", "bar")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList("foo", "bar")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.InList([]interface{}{})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList([]interface{}{})
		value.chain.assert(t, success)
		value.chain.clear()

		value.InList([]interface{}{"bar", "foo"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList([]interface{}{"bar", "foo"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.InList([]interface{}{"bar", "foo"}, []interface{}{"foo", "bar"})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotInList([]interface{}{"bar", "foo"}, []interface{}{"foo", "bar"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.InList([]interface{}{"bar", "foo"}, []interface{}{"FOO", "BAR"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList([]interface{}{"bar", "foo"}, []interface{}{"FOO", "BAR"})
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myArray []interface{}
			myMap   map[string]interface{}
			myInt   int
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{
			123,
			456,
			[]interface{}{789, 567},
			map[string]interface{}{"a": "b"},
		})

		value.InList(myArray{
			myInt(123.0),
			myInt(456.0),
			myArray{
				myInt(789.0),
				myInt(567.0),
			},
			myMap{"a": "b"},
		})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotInList(myArray{
			myInt(123.0),
			myInt(456.0),
			myArray{
				myInt(789.0),
				myInt(567.0),
			},
			myMap{"a": "b"},
		})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.InList(myArray{123.0, 456.0, myArray{789.0, 567.0}, myMap{"a": "b"}})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotInList(myArray{123.0, 456.0, myArray{789.0, 567.0}, myMap{"a": "b"}})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.InList(myArray{myInt(123), 456.0, myArray{myInt(789), 567.0}, myMap{"a": "b"}})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotInList(myArray{myInt(123), 456.0, myArray{myInt(789), 567.0}, myMap{"a": "b"}})
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("not array", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{"foo", "bar"})

		value.InList([]interface{}{"bar", "foo"}, "NOT ARRAY")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList([]interface{}{"bar", "foo"}, "NOT ARRAY")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.InList([]interface{}{"foo", "bar"}, "NOT ARRAY")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList([]interface{}{"foo", "bar"}, "NOT ARRAY")
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.InList()
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList()
		value.chain.assert(t, failure)
		value.chain.clear()

		value.InList(nil)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList(nil)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.InList(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotInList(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()
	})
}

func TestArray_ConsistsOf(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, "foo"})

		value.ConsistsOf(123)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotConsistsOf(123)
		value.chain.assert(t, success)
		value.chain.clear()

		value.ConsistsOf("foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotConsistsOf("foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.ConsistsOf("foo", 123)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotConsistsOf("foo", 123)
		value.chain.assert(t, success)
		value.chain.clear()

		value.ConsistsOf(123, "foo", "foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotConsistsOf(123, "foo", "foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.ConsistsOf(123, "foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotConsistsOf(123, "foo")
		value.chain.assert(t, failure)
		value.chain.clear()

	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt int
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456})

		assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

		value.ConsistsOf(myInt(123), 456.0)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotConsistsOf(myInt(123), 456.0)
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.ConsistsOf(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotConsistsOf(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()
	})
}

func TestArray_Contains(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, "foo"})

		value.Contains(123)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContains(123)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.Contains("foo", 123)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContains("foo", 123)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.Contains("foo", "foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContains("foo", "foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.Contains(123, "foo", "FOO")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContains(123, "foo", "FOO")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContains("FOO")
		value.chain.assert(t, success)
		value.chain.clear()

		value.Contains([]interface{}{123, "foo"})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContains([]interface{}{123, "foo"})
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt int
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456})

		assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

		value.Contains(myInt(123))
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContains(myInt(123))
		value.chain.assert(t, failure)
		value.chain.clear()

		value.Contains(456.0)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContains(456.0)
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.Contains(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContains(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()
	})
}

func TestArray_ContainsAll(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, "foo"})

		value.ContainsAll(123)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsAll(123)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsAll("foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsAll("foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsAll("FOO")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsAll("FOO")
		value.chain.assert(t, success)
		value.chain.clear()

		value.ContainsAll(123, "foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsAll(123, "foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsAll("foo", "foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsAll("foo", "foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsAll(123, "foo", "FOO")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsAll(123, "foo", "FOO")
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt int
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456})

		assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

		value.ContainsAll(myInt(123), 456.0)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsAll(myInt(123), 456.0)
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.ContainsAll(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsAll(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()
	})
}

func TestArray_ContainsAny(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, "foo"})

		value.ContainsAny(123)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsAny(123)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsAny("foo", 123)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsAny("foo", 123)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsAny("foo", "foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsAny("foo", "foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsAny(123, "foo", "FOO")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsAny(123, "foo", "FOO")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsAny("FOO")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsAny("FOO")
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt   int
			myArray []interface{}
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 789, "foo", []interface{}{567, 456}})

		value.ContainsAny(myInt(123), 789.0)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsAny(myInt(123), 789.0)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsAny(myArray{567.0, 456.0})
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsAny(myArray{567.0, 456.0})
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.ContainsAny(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsAny(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()
	})
}

func TestArray_ContainsOnly(t *testing.T) {
	t.Run("without duplicates", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, "foo"})

		value.ContainsOnly(123)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsOnly(123)
		value.chain.assert(t, success)
		value.chain.clear()

		value.ContainsOnly("foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsOnly("foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.ContainsOnly(123, "foo", "foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsOnly(123, "foo", "foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsOnly(123, "foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsOnly(123, "foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsOnly("foo", 123)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsOnly("foo", 123)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsOnly("foo", 123, "bar")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsOnly("foo", 123, "bar")
		value.chain.assert(t, success)
		value.chain.clear()
	})

	t.Run("with duplicates", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, "foo", "foo"})

		value.ContainsOnly(123, "foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsOnly(123, "foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsOnly(123, 123, "foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsOnly(123, 123, "foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsOnly(123, "foo", "foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsOnly(123, "foo", "foo")
		value.chain.assert(t, failure)
		value.chain.clear()

		value.ContainsOnly("foo", 123, "foo")
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsOnly("foo", 123, "foo")
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt int
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456, 456})

		value.ContainsOnly(456.0)
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsOnly(456.0)
		value.chain.assert(t, success)
		value.chain.clear()

		value.ContainsOnly(myInt(123), 456.0)
		value.chain.assert(t, success)
		value.chain.clear()

		value.NotContainsOnly(myInt(123), 456.0)
		value.chain.assert(t, failure)
		value.chain.clear()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.ContainsOnly(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()

		value.NotContainsOnly(func() {})
		value.chain.assert(t, failure)
		value.chain.clear()
	})
}

func TestArray_HasValue(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		array := NewArray(reporter, []interface{}{
			123,
			[]interface{}{"456", 789},
			map[string]interface{}{
				"a": "b",
			},
		})

		array.HasValue(0, 123)
		array.chain.assert(t, success)
		array.chain.clear()

		array.NotHasValue(0, 123)
		array.chain.assert(t, failure)
		array.chain.clear()

		array.HasValue(1, []interface{}{"456", 789})
		array.chain.assert(t, success)
		array.chain.clear()

		array.NotHasValue(1, []interface{}{"456", 789})
		array.chain.assert(t, failure)
		array.chain.clear()

		array.HasValue(2, map[string]interface{}{"a": "b"})
		array.chain.assert(t, success)
		array.chain.clear()

		array.NotHasValue(2, map[string]interface{}{"a": "b"})
		array.chain.assert(t, failure)
		array.chain.clear()

		array.HasValue(3, 777)
		array.chain.assert(t, failure)
		array.chain.clear()

		array.NotHasValue(3, 777)
		array.chain.assert(t, failure)
		array.chain.clear()
	})

	t.Run("struct", func(t *testing.T) {
		reporter := newMockReporter(t)

		array := NewArray(reporter, []interface{}{
			map[string]interface{}{
				"a": map[string]interface{}{
					"b": 333,
					"c": 444,
				},
			},
		})

		type (
			A struct {
				B int `json:"b"`
				C int `json:"c"`
			}

			Baz struct {
				A A `json:"a"`
			}
		)

		baz := Baz{
			A: A{
				B: 333,
				C: 444,
			},
		}

		array.HasValue(0, baz)
		array.chain.assert(t, success)
		array.chain.clear()

		array.NotHasValue(0, baz)
		array.chain.assert(t, failure)
		array.chain.clear()

		array.HasValue(0, Baz{})
		array.chain.assert(t, failure)
		array.chain.clear()

		array.NotHasValue(0, Baz{})
		array.chain.assert(t, success)
		array.chain.clear()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myArray []interface{}
			myMap   map[string]interface{}
			myInt   int
		)

		reporter := newMockReporter(t)

		array := NewArray(reporter, []interface{}{
			123,
			[]interface{}{"456", 789},
			map[string]interface{}{
				"a": "b",
			},
		})

		array.HasValue(1, myArray{"456", myInt(789)})
		array.chain.assert(t, success)
		array.chain.clear()

		array.NotHasValue(1, myArray{"456", myInt(789)})
		array.chain.assert(t, failure)
		array.chain.clear()

		array.HasValue(2, myMap{"a": "b"})
		array.chain.assert(t, success)
		array.chain.clear()

		array.NotHasValue(2, myMap{"a": "b"})
		array.chain.assert(t, failure)
		array.chain.clear()
	})

	t.Run("invalid index", func(t *testing.T) {
		reporter := newMockReporter(t)

		array := NewArray(reporter, []interface{}{1, 2, 3})

		array.HasValue(-1, 999)
		array.chain.assert(t, failure)
		array.chain.clear()

		array.NotHasValue(-1, 999)
		array.chain.assert(t, failure)
		array.chain.clear()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		array := NewArray(reporter, []interface{}{1, 2, 3})

		array.HasValue(1, func() {})
		array.chain.assert(t, failure)
		array.chain.clear()

		array.NotHasValue(1, func() {})
		array.chain.assert(t, failure)
		array.chain.clear()
	})
}

func TestArray_Every(t *testing.T) {
	t.Run("check value", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{2, 4, 6})

		invoked := 0
		array.Every(func(_ int, val *Value) {
			if v, ok := val.Raw().(float64); ok {
				invoked++
				assert.Equal(t, 0, int(v)%2)
			}
		})

		assert.Equal(t, 3, invoked)
		array.chain.assert(t, success)
	})

	t.Run("check index", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, 3})

		invoked := 0
		array.Every(func(idx int, val *Value) {
			if v, ok := val.Raw().(float64); ok {
				invoked++
				assert.Equal(t, idx, int(v)-1)
			}
		})

		assert.Equal(t, 3, invoked)
		array.chain.assert(t, success)
	})

	t.Run("empty array", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})

		invoked := 0
		array.Every(func(_ int, val *Value) {
			invoked++
		})

		assert.Equal(t, 0, invoked)
		array.chain.assert(t, success)
	})

	t.Run("one assertion fails", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{"foo", "", "bar"})

		invoked := 0
		array.Every(func(_ int, val *Value) {
			invoked++
			val.String().NotEmpty()
		})

		assert.Equal(t, 3, invoked)
		array.chain.assert(t, failure)
	})

	t.Run("all assertions fail", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{"", "", ""})

		invoked := 0
		array.Every(func(_ int, val *Value) {
			invoked++
			val.String().NotEmpty()
		})

		assert.Equal(t, 3, invoked)
		array.chain.assert(t, failure)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, 3})
		array.Every((func(index int, value *Value))(nil))
		array.chain.assert(t, failure)
	})
}

func TestArray_Transform(t *testing.T) {
	t.Run("check index", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, 3})

		newArray := array.Transform(func(idx int, val interface{}) interface{} {
			if v, ok := val.(float64); ok {
				assert.Equal(t, idx, int(v)-1)
			}
			return val
		})

		assert.Equal(t, []interface{}{float64(1), float64(2), float64(3)}, newArray.value)
		newArray.chain.assert(t, success)
	})

	t.Run("transform value", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{2, 4, 6})

		newArray := array.Transform(func(_ int, val interface{}) interface{} {
			if v, ok := val.(float64); ok {
				return int(v) * int(v)
			}
			t.Errorf("failed transformation")
			return nil
		})

		assert.Equal(t, []interface{}{float64(4), float64(16), float64(36)}, newArray.value)
		newArray.chain.assert(t, success)
	})

	t.Run("empty array", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})

		newArray := array.Transform(func(_ int, _ interface{}) interface{} {
			t.Errorf("failed transformation")
			return nil
		})

		newArray.chain.assert(t, success)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{2, 4, 6})

		newArray := array.Transform(nil)

		newArray.chain.assert(t, failure)
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt int
		)

		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{2, 4, 6})

		newArray := array.Transform(func(_ int, val interface{}) interface{} {
			if val, ok := val.(float64); ok {
				return myInt(val)
			}
			t.Errorf("failed transformation")
			return nil
		})

		assert.Equal(t, []interface{}{2.0, 4.0, 6.0}, newArray.Raw())
		newArray.chain.assert(t, success)
	})
}

func TestArray_Filter(t *testing.T) {
	t.Run("elements of same type", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, 3, 4, 5, 6})

		filteredArray := array.Filter(func(index int, value *Value) bool {
			return value.Raw() != 2.0 && value.Raw() != 5.0
		})

		assert.Equal(t, []interface{}{1.0, 3.0, 4.0, 6.0}, filteredArray.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, 2.0, 3.0, 4.0, 5.0, 6.0})

		array.chain.assert(t, success)
		filteredArray.chain.assert(t, success)
	})

	t.Run("elements of different types", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{"foo", "bar", true, 1.0})

		filteredArray := array.Filter(func(index int, value *Value) bool {
			return value.Raw() != "bar"
		})

		assert.Equal(t, []interface{}{"foo", true, 1.0}, filteredArray.Raw())
		assert.Equal(t, array.Raw(), []interface{}{"foo", "bar", true, 1.0})

		array.chain.assert(t, success)
		filteredArray.chain.assert(t, success)
	})

	t.Run("empty array", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})

		filteredArray := array.Filter(func(index int, value *Value) bool {
			return false
		})

		assert.Equal(t, []interface{}{}, filteredArray.Raw())
		assert.Equal(t, array.Raw(), []interface{}{})

		array.chain.assert(t, success)
		filteredArray.chain.assert(t, success)
	})

	t.Run("no match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{"foo", "bar", true, 1.0})

		filteredArray := array.Filter(func(index int, value *Value) bool {
			return false
		})

		assert.Equal(t, []interface{}{}, filteredArray.Raw())
		assert.Equal(t, array.Raw(), []interface{}{"foo", "bar", true, 1.0})

		array.chain.assert(t, success)
		filteredArray.chain.assert(t, success)
	})

	t.Run("assertion fails", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1.0, "foo", "bar", 4.0, "baz", 6.0})

		filteredArray := array.Filter(func(index int, value *Value) bool {
			stringifiedValue := value.String().NotEmpty().Raw()
			return stringifiedValue != "bar"
		})

		assert.Equal(t, []interface{}{"foo", "baz"}, filteredArray.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, "foo", "bar", 4.0, "baz", 6.0})

		array.chain.assert(t, success)
		filteredArray.chain.assert(t, success)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{"foo", "bar", true, 1.0})
		filteredArray := array.Filter((func(index int, value *Value) bool)(nil))
		array.chain.assert(t, failure)
		filteredArray.chain.assert(t, failure)
	})
}

func TestArray_Find(t *testing.T) {
	t.Run("elements of same type", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, 3, 4, 5, 6})

		foundValue := array.Find(func(index int, value *Value) bool {
			return value.Raw() == 2.0
		})

		assert.Equal(t, 2.0, foundValue.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, 2.0, 3.0, 4.0, 5.0, 6.0})

		array.chain.assert(t, success)
		foundValue.chain.assert(t, success)
	})

	t.Run("elements of different types", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, "foo", true, "bar"})

		foundValue := array.Find(func(index int, value *Value) bool {
			stringifiedValue := value.String().NotEmpty().Raw()
			return stringifiedValue == "bar"
		})

		assert.Equal(t, "bar", foundValue.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, "foo", true, "bar"})

		array.chain.assert(t, success)
		foundValue.chain.assert(t, success)
	})

	t.Run("first match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, "foo", true, "bar"})

		foundValue := array.Find(func(index int, value *Value) bool {
			stringifiedValue := value.String().Raw()
			return stringifiedValue != ""
		})

		assert.Equal(t, "foo", foundValue.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, "foo", true, "bar"})

		array.chain.assert(t, success)
		foundValue.chain.assert(t, success)
	})

	t.Run("no match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, "foo", true, "bar"})

		foundValue := array.Find(func(index int, value *Value) bool {
			return value.Raw() == 2.0
		})

		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, "foo", true, "bar"})

		array.chain.assert(t, failure)
		foundValue.chain.assert(t, failure)
	})

	t.Run("empty array", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})

		foundValue := array.Find(func(index int, value *Value) bool {
			return value.Raw() == 2.0
		})

		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, array.Raw(), []interface{}{})

		array.chain.assert(t, failure)
		foundValue.chain.assert(t, failure)
	})

	t.Run("predicate returns true, assertion fails, no match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2})

		foundValue := array.Find(func(index int, value *Value) bool {
			value.String()
			return true
		})

		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, 2.0})

		array.chain.assert(t, failure)
		foundValue.chain.assert(t, failure)
	})

	t.Run("predicate returns true, assertion fails, have match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, "str"})

		foundValue := array.Find(func(index int, value *Value) bool {
			value.String()
			return true
		})

		assert.Equal(t, "str", foundValue.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, 2.0, "str"})

		array.chain.assert(t, success)
		foundValue.chain.assert(t, success)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2})

		foundValue := array.Find(nil)

		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, 2.0})

		array.chain.assert(t, failure)
		foundValue.chain.assert(t, failure)
	})
}

func TestArray_FindAll(t *testing.T) {
	t.Run("elements of same type", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, 3, 4, 5, 6})

		foundValues := array.FindAll(func(index int, value *Value) bool {
			return value.Raw() == 2.0 || value.Raw() == 5.0
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}

		assert.Equal(t, []interface{}{2.0, 5.0}, actual)
		assert.Equal(t, array.Raw(), []interface{}{1.0, 2.0, 3.0, 4.0, 5.0, 6.0})

		array.chain.assert(t, success)
		for _, value := range foundValues {
			value.chain.assert(t, success)
		}
	})

	t.Run("elements of different types", func(t *testing.T) {
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
		assert.Equal(t, array.Raw(), []interface{}{1.0, "foo", true, "bar"})

		array.chain.assert(t, success)
		for _, value := range foundValues {
			value.chain.assert(t, success)
		}
	})

	t.Run("no match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1.0, "foo", true, "bar"})

		foundValues := array.FindAll(func(index int, value *Value) bool {
			return value.Raw() == 2.0
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}

		assert.Equal(t, []interface{}{}, actual)
		assert.Equal(t, array.Raw(), []interface{}{1.0, "foo", true, "bar"})

		array.chain.assert(t, success)
		for _, value := range foundValues {
			value.chain.assert(t, success)
		}
	})

	t.Run("empty array", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})

		foundValues := array.FindAll(func(index int, value *Value) bool {
			return value.Raw() == 2.0
		})

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}

		assert.Equal(t, []interface{}{}, actual)
		assert.Equal(t, array.Raw(), []interface{}{})

		array.chain.assert(t, success)
		for _, value := range foundValues {
			value.chain.assert(t, success)
		}
	})

	t.Run("predicate returns true, assertion fails, no match", func(t *testing.T) {
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
		assert.Equal(t, array.Raw(), []interface{}{1.0, 2.0})

		array.chain.assert(t, success)
		for _, value := range foundValues {
			value.chain.assert(t, success)
		}
	})

	t.Run("predicate returns true, assertion fails, have matches", func(t *testing.T) {
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
		assert.Equal(t, array.Raw(), []interface{}{"foo", 1.0, 2.0, "bar"})

		array.chain.assert(t, success)
		for _, value := range foundValues {
			value.chain.assert(t, success)
		}
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2})

		foundValues := array.FindAll(nil)

		actual := []interface{}{}
		for _, value := range foundValues {
			actual = append(actual, value.Raw())
		}

		assert.Equal(t, []interface{}{}, actual)
		assert.Equal(t, array.Raw(), []interface{}{1.0, 2.0})

		array.chain.assert(t, failure)
		for _, value := range foundValues {
			value.chain.assert(t, failure)
		}
	})
}

func TestArray_NotFind(t *testing.T) {
	t.Run("no match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, "foo", true, "bar"})

		afterArray := array.NotFind(func(index int, value *Value) bool {
			return value.String().Raw() == "baz"
		})

		assert.Same(t, array, afterArray)
		array.chain.assert(t, success)
	})

	t.Run("have match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, "foo", true, "bar"})

		afterArray := array.NotFind(func(index int, value *Value) bool {
			return value.String().NotEmpty().Raw() == "bar"
		})

		assert.Same(t, array, afterArray)
		array.chain.assert(t, failure)
	})

	t.Run("empty array", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})

		afterArray := array.NotFind(func(index int, value *Value) bool {
			return value.Raw() == 2.0
		})

		assert.Same(t, array, afterArray)
		array.chain.assert(t, success)
	})

	t.Run("predicate returns true, assertion fails, no match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2})

		afterArray := array.NotFind(func(index int, value *Value) bool {
			value.String()
			return true
		})

		assert.Same(t, array, afterArray)
		array.chain.assert(t, success)
	})

	t.Run("predicate returns true, assertion fails, have match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, "str"})

		afterArray := array.NotFind(func(index int, value *Value) bool {
			value.String()
			return true
		})

		assert.Same(t, array, afterArray)
		array.chain.assert(t, failure)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2})

		afterArray := array.NotFind(nil)

		assert.Same(t, array, afterArray)
		array.chain.assert(t, failure)
	})
}

func TestArray_IsOrdered(t *testing.T) {
	type args struct {
		values      []interface{}
		less        []func(x, y *Value) bool
		chainFailed bool
	}
	cases := []struct {
		name        string
		args        args
		isInvalid   bool
		isOrdered   bool
		isUnordered bool
	}{
		{
			name: "booleans ordered",
			args: args{
				values: []interface{}{false, false, true, true},
			},
			isInvalid:   false,
			isOrdered:   true,
			isUnordered: false,
		},
		{
			name: "booleans unordered",
			args: args{
				values: []interface{}{true, false, true, false},
			},
			isInvalid:   false,
			isOrdered:   false,
			isUnordered: true,
		},
		{
			name: "numbers ordered",
			args: args{
				values: []interface{}{1, 1, 2, 3},
			},
			isInvalid:   false,
			isOrdered:   true,
			isUnordered: false,
		},
		{
			name: "numbers unordered",
			args: args{
				values: []interface{}{3, 1, 1, 2},
			},
			isInvalid:   false,
			isOrdered:   false,
			isUnordered: true,
		},
		{
			name: "strings ordered",
			args: args{
				values: []interface{}{"", "a", "b", "ba"},
			},
			isInvalid:   false,
			isOrdered:   true,
			isUnordered: false,
		},
		{
			name: "strings unordered",
			args: args{
				values: []interface{}{"z", "y", "x", ""},
			},
			isInvalid:   false,
			isOrdered:   false,
			isUnordered: true,
		},
		{
			name: "all nils",
			args: args{
				values: []interface{}{nil, nil, nil},
			},
			isInvalid:   false,
			isOrdered:   true,
			isUnordered: false,
		},
		{
			name: "reversed",
			args: args{
				values: []interface{}{3, 2, 1},
			},
			isInvalid:   false,
			isOrdered:   false,
			isUnordered: true,
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
			isInvalid:   false,
			isOrdered:   true,
			isUnordered: false,
		},
		{
			name: "user-defined less function, negated",
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
			isInvalid:   false,
			isOrdered:   false,
			isUnordered: true,
		},
		{
			name: "empty array",
			args: args{
				values: []interface{}{},
			},
			isInvalid:   false,
			isOrdered:   true,
			isUnordered: true,
		},
		{
			name: "one element",
			args: args{
				values: []interface{}{1},
			},
			isInvalid:   false,
			isOrdered:   true,
			isUnordered: true,
		},
		{
			name: "empty array, custom func",
			args: args{
				values: []interface{}{},
				less: []func(x, y *Value) bool{
					func(x, y *Value) bool {
						panic("test")
					},
				},
			},
			isInvalid:   false,
			isOrdered:   true,
			isUnordered: true,
		},
		{
			name: "one element, custom func",
			args: args{
				values: []interface{}{1},
				less: []func(x, y *Value) bool{
					func(x, y *Value) bool {
						panic("test")
					},
				},
			},
			isInvalid:   false,
			isOrdered:   true,
			isUnordered: true,
		},
		{
			name: "invalid, assertion failed",
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
			isInvalid: true,
		},
		{
			name: "invalid, multiple arguments",
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
			isInvalid: true,
		},
		{
			name: "invalid, nil argument",
			args: args{
				values: []interface{}{1, 2, 3},
				less: []func(x, y *Value) bool{
					nil,
				},
			},
			isInvalid: true,
		},
		{
			name: "invalid, unsupported type",
			args: args{
				values: []interface{}{[]int{1, 2}, []int{3, 4}, []int{5, 6}},
				less:   []func(x, y *Value) bool{},
			},
			isInvalid: true,
		},
		{
			name: "invalid, multiple types",
			args: args{
				values: []interface{}{1, "abc", true},
				less:   []func(x, y *Value) bool{},
			},
			isInvalid: true,
		},
		{
			name: "invalid, failed chain",
			args: args{
				chainFailed: true,
			},
			isInvalid: true,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if tc.isInvalid {
				t.Run("normal", func(t *testing.T) {
					NewArray(reporter, tc.args.values).IsOrdered(tc.args.less...).
						chain.assert(t, failure)

					NewArray(reporter, tc.args.values).NotOrdered(tc.args.less...).
						chain.assert(t, failure)
				})

				t.Run("reversed", func(t *testing.T) {
					// reverse slice
					sort.SliceStable(tc.args.values, func(i, j int) bool {
						return i > j
					})

					NewArray(reporter, tc.args.values).IsOrdered(tc.args.less...).
						chain.assert(t, failure)

					NewArray(reporter, tc.args.values).NotOrdered(tc.args.less...).
						chain.assert(t, failure)
				})
			} else {
				t.Run("is ordered", func(t *testing.T) {
					if tc.isOrdered {
						NewArray(reporter, tc.args.values).IsOrdered(tc.args.less...).
							chain.assert(t, success)
					} else {
						NewArray(reporter, tc.args.values).IsOrdered(tc.args.less...).
							chain.assert(t, failure)
					}
				})

				t.Run("not ordered", func(t *testing.T) {
					if tc.isUnordered {
						NewArray(reporter, tc.args.values).NotOrdered(tc.args.less...).
							chain.assert(t, success)
					} else {
						NewArray(reporter, tc.args.values).NotOrdered(tc.args.less...).
							chain.assert(t, failure)
					}
				})
			}
		})
	}
}

func TestArray_ComparatorErrors(t *testing.T) {
	t.Run("nil slice", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, nil)

		assert.Nil(t, fn)
		chain.assert(t, success)
	})

	t.Run("0 elements", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, []interface{}{})

		assert.Nil(t, fn)
		chain.assert(t, success)
	})

	t.Run("1 element", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, []interface{}{
			"test",
		})

		assert.Nil(t, fn)
		chain.assert(t, success)
	})

	t.Run("2 elements, bad_type", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, []interface{}{
			"test",
			make(chan int), // bad type
		})

		assert.Nil(t, fn)
		chain.assert(t, failure)
	})

	t.Run("2 elements, good types", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, []interface{}{
			"test",
			"test",
		})

		assert.NotNil(t, fn)
		chain.assert(t, success)
	})
}
