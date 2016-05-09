package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestObjectFailed(t *testing.T) {
	checker := newMockChecker(t)

	checker.Fail("fail")

	value := NewObject(checker, nil)

	assert.False(t, value.Keys() == nil)
	assert.False(t, value.Values() == nil)
	assert.False(t, value.Value("foo") == nil)

	value.Empty()
	value.NotEmpty()
	value.Equal(nil)
	value.NotEqual(nil)
	value.ContainsKey("foo")
	value.NotContainsKey("foo")
	value.ContainsMap(nil)
	value.NotContainsMap(nil)
	value.ValueEqual("foo", nil)
	value.ValueNotEqual("foo", nil)
}

func TestObjectGetters(t *testing.T) {
	checker := newMockChecker(t)

	m := map[string]interface{}{
		"foo": 123.0,
		"bar": []interface{}{"456", 789.0},
		"baz": map[string]interface{}{
			"a": "b",
		},
	}

	value := NewObject(checker, m)

	keys := []interface{}{"foo", "bar", "baz"}

	values := []interface{}{
		123.0,
		[]interface{}{"456", 789.0},
		map[string]interface{}{
			"a": "b",
		},
	}

	value.Keys().ContainsOnly(keys...)
	checker.AssertSuccess(t)
	checker.Reset()

	value.Values().ContainsOnly(values...)
	checker.AssertSuccess(t)
	checker.Reset()

	assert.Equal(t, m["foo"], value.Value("foo").Raw())
	checker.AssertSuccess(t)
	checker.Reset()

	assert.Equal(t, m["bar"], value.Value("bar").Raw())
	checker.AssertSuccess(t)
	checker.Reset()

	assert.Equal(t, m["baz"], value.Value("baz").Raw())
	checker.AssertSuccess(t)
	checker.Reset()

	assert.Equal(t, nil, value.Value("BAZ").Raw())
	checker.AssertFailed(t)
	checker.Reset()

	assert.False(t, value.checker == value.Keys().checker)
	assert.False(t, value.checker == value.Values().checker)
	assert.False(t, value.checker == value.Value("foo").checker)
}

func TestObjectEmpty(t *testing.T) {
	checker := newMockChecker(t)

	value1 := NewObject(checker, nil)

	_ = value1
	checker.AssertFailed(t)
	checker.Reset()

	value2 := NewObject(checker, map[string]interface{}{})

	value2.Empty()
	checker.AssertSuccess(t)
	checker.Reset()

	value2.NotEmpty()
	checker.AssertFailed(t)
	checker.Reset()

	value3 := NewObject(checker, map[string]interface{}{"": nil})

	value3.Empty()
	checker.AssertFailed(t)
	checker.Reset()

	value3.NotEmpty()
	checker.AssertSuccess(t)
	checker.Reset()
}

func TestObjectEqualEmpty(t *testing.T) {
	checker := newMockChecker(t)

	value := NewObject(checker, map[string]interface{}{})

	assert.Equal(t, map[string]interface{}{}, value.Raw())

	value.Equal(map[string]interface{}{})
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual(map[string]interface{}{})
	checker.AssertFailed(t)
	checker.Reset()

	value.Equal(map[string]interface{}{"": nil})
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual(map[string]interface{}{"": nil})
	checker.AssertSuccess(t)
	checker.Reset()
}

func TestObjectEqual(t *testing.T) {
	checker := newMockChecker(t)

	value := NewObject(checker, map[string]interface{}{"foo": 123.0})

	assert.Equal(t, map[string]interface{}{"foo": 123.0}, value.Raw())

	value.Equal(map[string]interface{}{})
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual(map[string]interface{}{})
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal(map[string]interface{}{"FOO": 123.0})
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual(map[string]interface{}{"FOO": 123.0})
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal(map[string]interface{}{"foo": 456.0})
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual(map[string]interface{}{"foo": 456.0})
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal(map[string]interface{}{"foo": 123.0})
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual(map[string]interface{}{"foo": 123.0})
	checker.AssertFailed(t)
	checker.Reset()

	value.Equal(nil)
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual(nil)
	checker.AssertFailed(t)
	checker.Reset()
}

func TestObjectContainsKey(t *testing.T) {
	checker := newMockChecker(t)

	value := NewObject(checker, map[string]interface{}{"foo": 123, "bar": ""})

	value.ContainsKey("foo")
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotContainsKey("foo")
	checker.AssertFailed(t)
	checker.Reset()

	value.ContainsKey("bar")
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotContainsKey("bar")
	checker.AssertFailed(t)
	checker.Reset()

	value.ContainsKey("BAR")
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContainsKey("BAR")
	checker.AssertSuccess(t)
	checker.Reset()
}

func TestObjectContainsMapSuccess(t *testing.T) {
	checker := newMockChecker(t)

	value := NewObject(checker, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": map[string]interface{}{
				"b": 333,
				"c": 444,
			},
		},
	})

	submap1 := map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
	}

	value.ContainsMap(submap1)
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotContainsMap(submap1)
	checker.AssertFailed(t)
	checker.Reset()

	submap2 := map[string]interface{}{
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": map[string]interface{}{
				"c": 444,
			},
		},
	}

	value.ContainsMap(submap2)
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotContainsMap(submap2)
	checker.AssertFailed(t)
	checker.Reset()
}

func TestObjectContainsMapFailed(t *testing.T) {
	checker := newMockChecker(t)

	value := NewObject(checker, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": map[string]interface{}{
				"b": 333,
				"c": 444,
			},
		},
	})

	submap1 := map[string]interface{}{
		"foo": 123,
		"qux": 456,
	}

	value.ContainsMap(submap1)
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContainsMap(submap1)
	checker.AssertSuccess(t)
	checker.Reset()

	submap2 := map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", "789"},
	}

	value.ContainsMap(submap2)
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContainsMap(submap2)
	checker.AssertSuccess(t)
	checker.Reset()

	submap3 := map[string]interface{}{
		"baz": map[string]interface{}{
			"a": map[string]interface{}{
				"b": "333",
				"c": 444,
			},
		},
	}

	value.ContainsMap(submap3)
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContainsMap(submap3)
	checker.AssertSuccess(t)
	checker.Reset()

	value.ContainsMap(nil)
	checker.AssertFailed(t)
	checker.Reset()

	value.NotContainsMap(nil)
	checker.AssertFailed(t)
	checker.Reset()
}

func TestObjectValueEqual(t *testing.T) {
	checker := newMockChecker(t)

	value := NewObject(checker, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": "b",
		},
	})

	value.ValueEqual("foo", 123)
	checker.AssertSuccess(t)
	checker.Reset()

	value.ValueNotEqual("foo", 123)
	checker.AssertFailed(t)
	checker.Reset()

	value.ValueEqual("bar", []interface{}{"456", 789})
	checker.AssertSuccess(t)
	checker.Reset()

	value.ValueNotEqual("bar", []interface{}{"456", 789})
	checker.AssertFailed(t)
	checker.Reset()

	value.ValueEqual("baz", map[string]interface{}{"a": "b"})
	checker.AssertSuccess(t)
	checker.Reset()

	value.ValueNotEqual("baz", map[string]interface{}{"a": "b"})
	checker.AssertFailed(t)
	checker.Reset()

	value.ValueEqual("baz", func() {})
	checker.AssertFailed(t)
	checker.Reset()

	value.ValueNotEqual("baz", func() {})
	checker.AssertFailed(t)
	checker.Reset()

	value.ValueEqual("BAZ", 777)
	checker.AssertFailed(t)
	checker.Reset()

	value.ValueNotEqual("BAZ", 777)
	checker.AssertFailed(t)
	checker.Reset()
}

func TestObjectConvertEqual(t *testing.T) {
	type (
		myMap map[string]interface{}
		myInt int
	)

	checker := newMockChecker(t)

	value := NewObject(checker, map[string]interface{}{"foo": 123})

	value.Equal(map[string]interface{}{"foo": "123"})
	checker.AssertFailed(t)
	checker.Reset()

	value.NotEqual(map[string]interface{}{"foo": "123"})
	checker.AssertSuccess(t)
	checker.Reset()

	value.Equal(map[string]interface{}{"foo": 123.0})
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual(map[string]interface{}{"foo": 123.0})
	checker.AssertFailed(t)
	checker.Reset()

	value.Equal(map[string]interface{}{"foo": 123})
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual(map[string]interface{}{"foo": 123})
	checker.AssertFailed(t)
	checker.Reset()

	value.Equal(myMap{"foo": myInt(123)})
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotEqual(myMap{"foo": myInt(123)})
	checker.AssertFailed(t)
	checker.Reset()
}

func TestObjectConvertContainsMap(t *testing.T) {
	type (
		myArray []interface{}
		myMap   map[string]interface{}
		myInt   int
	)

	checker := newMockChecker(t)

	value := NewObject(checker, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": "b",
		},
	})

	submap := myMap{
		"foo": myInt(123),
		"bar": myArray{"456", myInt(789)},
	}

	value.ContainsMap(submap)
	checker.AssertSuccess(t)
	checker.Reset()

	value.NotContainsMap(submap)
	checker.AssertFailed(t)
	checker.Reset()
}

func TestObjectConvertValueEqual(t *testing.T) {
	type (
		myArray []interface{}
		myMap   map[string]interface{}
		myInt   int
	)

	checker := newMockChecker(t)

	value := NewObject(checker, map[string]interface{}{
		"foo": 123,
		"bar": []interface{}{"456", 789},
		"baz": map[string]interface{}{
			"a": "b",
		},
	})

	value.ValueEqual("bar", myArray{"456", myInt(789)})
	checker.AssertSuccess(t)
	checker.Reset()

	value.ValueNotEqual("bar", myArray{"456", myInt(789)})
	checker.AssertFailed(t)
	checker.Reset()

	value.ValueEqual("baz", myMap{"a": "b"})
	checker.AssertSuccess(t)
	checker.Reset()

	value.ValueNotEqual("baz", myMap{"a": "b"})
	checker.AssertFailed(t)
	checker.Reset()
}
