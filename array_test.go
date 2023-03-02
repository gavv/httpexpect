package httpexpect

import (
	"sort"
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
		value.Value(0).chain.assertFailed(t)
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
		chain := newMockChain(t)
		chain.setFailed()

		value := newArray(chain, []interface{}{})

		check(value)
	})

	t.Run("nil value", func(t *testing.T) {
		chain := newMockChain(t)

		value := newArray(chain, nil)

		check(value)
	})

	t.Run("failed chain, nil value", func(t *testing.T) {
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

	t.Run("invalid value", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, nil)

		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestArray_Decode(t *testing.T) {
	t.Run("target is empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		testValue := []interface{}{"Foo", 123.0}
		arr := NewArray(reporter, testValue)

		var target interface{}
		arr.Decode(&target)

		arr.chain.assertNotFailed(t)
		assert.Equal(t, testValue, target)
	})

	t.Run("target is slice of empty interfaces", func(t *testing.T) {
		reporter := newMockReporter(t)

		testValue := []interface{}{"Foo", 123.0}
		arr := NewArray(reporter, testValue)

		var target []interface{}
		arr.Decode(&target)

		arr.chain.assertNotFailed(t)
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

		arr.chain.assertNotFailed(t)
		assert.Equal(t, actualStruct, target)
	})

	t.Run("target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		testValue := []interface{}{"Foo", 123.0}
		arr := NewArray(reporter, testValue)

		arr.Decode(123)

		arr.chain.assertFailed(t)
	})

	t.Run("target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		testValue := []interface{}{"Foo", 123.0}
		arr := NewArray(reporter, testValue)

		arr.Decode(nil)

		arr.chain.assertFailed(t)
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
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		assert.Equal(t, data, value.Path("$").Raw())
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.Schema(`{"type": "array"}`)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.Schema(`{"type": "object"}`)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		assert.Equal(t, 0.0, value.Length().Raw())
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		assert.NotNil(t, value.Value(0))
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
	})

	t.Run("not empty", func(t *testing.T) {
		reporter := newMockReporter(t)

		data := []interface{}{"foo", 123.0}

		value := NewArray(reporter, data)

		assert.Equal(t, data, value.Raw())
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		assert.Equal(t, data, value.Path("$").Raw())
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.Schema(`{"type": "array"}`)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.Schema(`{"type": "object"}`)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		assert.Equal(t, 2.0, value.Length().Raw())
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		assert.Equal(t, "foo", value.Value(0).Raw())
		assert.Equal(t, 123.0, value.Value(1).Raw())
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		assert.Equal(t, nil, value.Value(2).Raw())
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		assert.Equal(t, "foo", value.First().Raw())
		assert.Equal(t, 123.0, value.Last().Raw())
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		it := value.Iter()
		assert.Equal(t, 2, len(it))
		assert.Equal(t, "foo", it[0].Raw())
		assert.Equal(t, 123.0, it[1].Raw())
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()
	})
}

func TestArray_IsEmpty(t *testing.T) {
	t.Run("empty slice", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.IsEmpty()
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEmpty()
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("one empty element", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{""})

		value.IsEmpty()
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEmpty()
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()
	})
}

func TestArray_IsEqual(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
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
	})

	t.Run("not empty", func(t *testing.T) {
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
	})

	t.Run("strings", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{"foo", "bar"})

		value.IsEqual([]string{"foo", "bar"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.IsEqual([]string{"bar", "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual([]string{"foo", "bar"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual([]string{"bar", "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()
	})

	t.Run("numbers", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456})

		value.IsEqual([]int{123, 456})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.IsEqual([]int{456, 123})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual([]int{123, 456})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual([]int{456, 123})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()
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
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.IsEqual([]S{{456}, {123}})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual([]S{{123}, {456}})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual([]S{{456}, {123}})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()
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
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqual(myArray{myInt(123), 456.0})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.IsEqual(nil)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual(nil)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsEqual(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqual(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestArray_IsEqualUnordered(t *testing.T) {
	t.Run("without duplicates", func(t *testing.T) {
		reporter := newMockReporter(t)

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

	t.Run("with duplicates", func(t *testing.T) {
		reporter := newMockReporter(t)

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

	t.Run("canonization", func(t *testing.T) {
		type (
			myArray []interface{}
			myInt   int
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456, "foo"})

		assert.Equal(t, []interface{}{123.0, 456.0, "foo"}, value.Raw())

		value.IsEqualUnordered(myArray{myInt(456), 123.0, "foo"})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered(myArray{myInt(456), 123.0, "foo"})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.IsEqualUnordered(nil)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered(nil)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.IsEqualUnordered(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotEqualUnordered(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestArray_InList(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{"foo", "bar"})

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
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotInList(myArray{
			myInt(123.0),
			myInt(456.0),
			myArray{
				myInt(789.0),
				myInt(567.0),
			},
			myMap{"a": "b"},
		})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.InList(myArray{123.0, 456.0, myArray{789.0, 567.0}, myMap{"a": "b"}})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotInList(myArray{123.0, 456.0, myArray{789.0, 567.0}, myMap{"a": "b"}})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.InList(myArray{myInt(123), 456.0, myArray{myInt(789), 567.0}, myMap{"a": "b"}})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotInList(myArray{myInt(123), 456.0, myArray{myInt(789), 567.0}, myMap{"a": "b"}})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("not array", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{"foo", "bar"})

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
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.InList()
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotInList()
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.InList(nil)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotInList(nil)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.InList(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotInList(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestArray_ConsistsOf(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, "foo"})

		value.ConsistsOf(123)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotConsistsOf(123)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.ConsistsOf("foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotConsistsOf("foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.ConsistsOf("foo", 123)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotConsistsOf("foo", 123)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.ConsistsOf(123, "foo", "foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotConsistsOf(123, "foo", "foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.ConsistsOf(123, "foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotConsistsOf(123, "foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt int
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456})

		assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

		value.ConsistsOf(myInt(123), 456.0)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotConsistsOf(myInt(123), 456.0)
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.ConsistsOf(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotConsistsOf(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestArray_Contains(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
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
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt int
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456})

		assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

		value.Contains(myInt(123))
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContains(myInt(123))
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.Contains(456.0)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContains(456.0)
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.Contains(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotContains(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestArray_ContainsAll(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, "foo"})

		value.ContainsAll(123)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsAll(123)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ContainsAll("foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsAll("foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ContainsAll("FOO")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotContainsAll("FOO")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.ContainsAll(123, "foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsAll(123, "foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ContainsAll("foo", "foo")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsAll("foo", "foo")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ContainsAll(123, "foo", "FOO")
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotContainsAll(123, "foo", "FOO")
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt int
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456})

		assert.Equal(t, []interface{}{123.0, 456.0}, value.Raw())

		value.ContainsAll(myInt(123), 456.0)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsAll(myInt(123), 456.0)
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.ContainsAll(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotContainsAll(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestArray_ContainsAny(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
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
	})

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt   int
			myArray []interface{}
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 789, "foo", []interface{}{567, 456}})

		value.ContainsAny(myInt(123), 789.0)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsAny(myInt(123), 789.0)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.ContainsAny(myArray{567.0, 456.0})
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsAny(myArray{567.0, 456.0})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.ContainsAny(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotContainsAny(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})
}

func TestArray_ContainsOnly(t *testing.T) {
	t.Run("without duplicates", func(t *testing.T) {
		reporter := newMockReporter(t)

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

	t.Run("with duplicates", func(t *testing.T) {
		reporter := newMockReporter(t)

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

	t.Run("canonization", func(t *testing.T) {
		type (
			myInt int
		)

		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{123, 456, 456})

		value.ContainsOnly(456.0)
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotContainsOnly(456.0)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.ContainsOnly(myInt(123), 456.0)
		value.chain.assertNotFailed(t)
		value.chain.clearFailed()

		value.NotContainsOnly(myInt(123), 456.0)
		value.chain.assertFailed(t)
		value.chain.clearFailed()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewArray(reporter, []interface{}{})

		value.ContainsOnly(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()

		value.NotContainsOnly(func() {})
		value.chain.assertFailed(t)
		value.chain.clearFailed()
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
		array.chain.assertNotFailed(t)
		array.chain.clearFailed()

		array.NotHasValue(0, 123)
		array.chain.assertFailed(t)
		array.chain.clearFailed()

		array.HasValue(1, []interface{}{"456", 789})
		array.chain.assertNotFailed(t)
		array.chain.clearFailed()

		array.NotHasValue(1, []interface{}{"456", 789})
		array.chain.assertFailed(t)
		array.chain.clearFailed()

		array.HasValue(2, map[string]interface{}{"a": "b"})
		array.chain.assertNotFailed(t)
		array.chain.clearFailed()

		array.NotHasValue(2, map[string]interface{}{"a": "b"})
		array.chain.assertFailed(t)
		array.chain.clearFailed()

		array.HasValue(3, 777)
		array.chain.assertFailed(t)
		array.chain.clearFailed()

		array.NotHasValue(3, 777)
		array.chain.assertFailed(t)
		array.chain.clearFailed()
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
		array.chain.assertNotFailed(t)
		array.chain.clearFailed()

		array.NotHasValue(0, baz)
		array.chain.assertFailed(t)
		array.chain.clearFailed()

		array.HasValue(0, Baz{})
		array.chain.assertFailed(t)
		array.chain.clearFailed()

		array.NotHasValue(0, Baz{})
		array.chain.assertNotFailed(t)
		array.chain.clearFailed()
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
		array.chain.assertNotFailed(t)
		array.chain.clearFailed()

		array.NotHasValue(1, myArray{"456", myInt(789)})
		array.chain.assertFailed(t)
		array.chain.clearFailed()

		array.HasValue(2, myMap{"a": "b"})
		array.chain.assertNotFailed(t)
		array.chain.clearFailed()

		array.NotHasValue(2, myMap{"a": "b"})
		array.chain.assertFailed(t)
		array.chain.clearFailed()
	})

	t.Run("invalid index", func(t *testing.T) {
		reporter := newMockReporter(t)

		array := NewArray(reporter, []interface{}{1, 2, 3})

		array.HasValue(-1, 999)
		array.chain.assertFailed(t)
		array.chain.clearFailed()

		array.NotHasValue(-1, 999)
		array.chain.assertFailed(t)
		array.chain.clearFailed()
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		array := NewArray(reporter, []interface{}{1, 2, 3})

		array.HasValue(1, func() {})
		array.chain.assertFailed(t)
		array.chain.clearFailed()

		array.NotHasValue(1, func() {})
		array.chain.assertFailed(t)
		array.chain.clearFailed()
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
		array.chain.assertNotFailed(t)
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
		array.chain.assertNotFailed(t)
	})

	t.Run("empty array", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})

		invoked := 0
		array.Every(func(_ int, val *Value) {
			invoked++
		})

		assert.Equal(t, 0, invoked)
		array.chain.assertNotFailed(t)
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
		array.chain.assertFailed(t)
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
		array.chain.assertFailed(t)
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
		newArray.chain.assertNotFailed(t)
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
		newArray.chain.assertNotFailed(t)
	})

	t.Run("empty array", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})

		newArray := array.Transform(func(_ int, _ interface{}) interface{} {
			t.Errorf("failed transformation")
			return nil
		})

		newArray.chain.assertNotFailed(t)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{2, 4, 6})

		newArray := array.Transform(nil)

		newArray.chain.assertFailed(t)
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
		newArray.chain.assertNotFailed(t)
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

		array.chain.assertNotFailed(t)
		filteredArray.chain.assertNotFailed(t)
	})

	t.Run("elements of different types", func(t *testing.T) {
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

	t.Run("empty array", func(t *testing.T) {
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

	t.Run("no match", func(t *testing.T) {
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

	t.Run("assertion fails", func(t *testing.T) {
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

		array.chain.assertNotFailed(t)
		foundValue.chain.assertNotFailed(t)
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

		array.chain.assertNotFailed(t)
		foundValue.chain.assertNotFailed(t)
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

		array.chain.assertNotFailed(t)
		foundValue.chain.assertNotFailed(t)
	})

	t.Run("no match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, "foo", true, "bar"})

		foundValue := array.Find(func(index int, value *Value) bool {
			return value.Raw() == 2.0
		})

		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, "foo", true, "bar"})

		array.chain.assertFailed(t)
		foundValue.chain.assertFailed(t)
	})

	t.Run("empty array", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})

		foundValue := array.Find(func(index int, value *Value) bool {
			return value.Raw() == 2.0
		})

		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, array.Raw(), []interface{}{})

		array.chain.assertFailed(t)
		foundValue.chain.assertFailed(t)
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

		array.chain.assertFailed(t)
		foundValue.chain.assertFailed(t)
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

		array.chain.assertNotFailed(t)
		foundValue.chain.assertNotFailed(t)
	})

	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2})

		foundValue := array.Find(nil)

		assert.Equal(t, nil, foundValue.Raw())
		assert.Equal(t, array.Raw(), []interface{}{1.0, 2.0})

		array.chain.assertFailed(t)
		foundValue.chain.assertFailed(t)
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

		array.chain.assertNotFailed(t)
		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
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

		array.chain.assertNotFailed(t)
		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
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

		array.chain.assertNotFailed(t)
		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
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

		array.chain.assertNotFailed(t)
		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
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

		array.chain.assertNotFailed(t)
		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
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

		array.chain.assertNotFailed(t)
		for _, value := range foundValues {
			value.chain.assertNotFailed(t)
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

		array.chain.assertFailed(t)
		for _, value := range foundValues {
			value.chain.assertFailed(t)
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
		array.chain.assertNotFailed(t)
	})

	t.Run("have match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, "foo", true, "bar"})

		afterArray := array.NotFind(func(index int, value *Value) bool {
			return value.String().NotEmpty().Raw() == "bar"
		})

		assert.Same(t, array, afterArray)
		array.chain.assertFailed(t)
	})

	t.Run("empty array", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{})

		afterArray := array.NotFind(func(index int, value *Value) bool {
			return value.Raw() == 2.0
		})

		assert.Same(t, array, afterArray)
		array.chain.assertNotFailed(t)
	})

	t.Run("predicate returns true, assertion fails, no match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2})

		afterArray := array.NotFind(func(index int, value *Value) bool {
			value.String()
			return true
		})

		assert.Same(t, array, afterArray)
		array.chain.assertNotFailed(t)
	})

	t.Run("predicate returns true, assertion fails, have match", func(t *testing.T) {
		reporter := newMockReporter(t)
		array := NewArray(reporter, []interface{}{1, 2, "str"})

		afterArray := array.NotFind(func(index int, value *Value) bool {
			value.String()
			return true
		})

		assert.Same(t, array, afterArray)
		array.chain.assertFailed(t)
	})

	t.Run("invalid argument", func(t *testing.T) {
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
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if tt.isInvalid {
				t.Run("normal", func(t *testing.T) {
					NewArray(reporter, tt.args.values).IsOrdered(tt.args.less...).
						chain.assertFailed(t)

					NewArray(reporter, tt.args.values).NotOrdered(tt.args.less...).
						chain.assertFailed(t)
				})

				t.Run("reversed", func(t *testing.T) {
					// reverse slice
					sort.SliceStable(tt.args.values, func(i, j int) bool {
						return i > j
					})

					NewArray(reporter, tt.args.values).IsOrdered(tt.args.less...).
						chain.assertFailed(t)

					NewArray(reporter, tt.args.values).NotOrdered(tt.args.less...).
						chain.assertFailed(t)
				})
			} else {
				t.Run("is ordered", func(t *testing.T) {
					if tt.isOrdered {
						NewArray(reporter, tt.args.values).IsOrdered(tt.args.less...).
							chain.assertNotFailed(t)
					} else {
						NewArray(reporter, tt.args.values).IsOrdered(tt.args.less...).
							chain.assertFailed(t)
					}
				})

				t.Run("not ordered", func(t *testing.T) {
					if tt.isUnordered {
						NewArray(reporter, tt.args.values).NotOrdered(tt.args.less...).
							chain.assertNotFailed(t)
					} else {
						NewArray(reporter, tt.args.values).NotOrdered(tt.args.less...).
							chain.assertFailed(t)
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
		chain.assertNotFailed(t)
	})

	t.Run("0 elements", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, []interface{}{})

		assert.Nil(t, fn)
		chain.assertNotFailed(t)
	})

	t.Run("1 element", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, []interface{}{
			"test",
		})

		assert.Nil(t, fn)
		chain.assertNotFailed(t)
	})

	t.Run("2 elements, bad_type", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, []interface{}{
			"test",
			make(chan int), // bad type
		})

		assert.Nil(t, fn)
		chain.assertFailed(t)
	})

	t.Run("2 elements, good types", func(t *testing.T) {
		chain := newMockChain(t).enter("test")

		fn := builtinComparator(chain, []interface{}{
			"test",
			"test",
		})

		assert.NotNil(t, fn)
		chain.assertNotFailed(t)
	})
}
