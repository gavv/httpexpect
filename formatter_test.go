package httpexpect

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type typedNil int

func (*typedNil) String() string {
	return ""
}

func TestFormatValues(t *testing.T) {
	checkAll := func(t *testing.T, fn func(interface{}) string) {
		var tnil *typedNil
		var tnilPtr fmt.Stringer = tnil

		assert.Nil(t, tnilPtr)
		assert.NotEqual(t, nil, tnilPtr)

		check := func(s string) {
			t.Logf("\n%s", s)
			assert.NotEmpty(t, s)
		}

		check(fn(nil))
		check(fn(tnil))
		check(fn(tnilPtr))
		check(fn(123))
		check(fn("hello"))
		check(fn(time.Second))
		check(fn(time.Unix(0, 0)))
		check(fn([]interface{}{1, 2}))
		check(fn(map[string]string{"a": "b"}))
		check(fn(make(chan int)))
		check(fn(AssertionRange{1, 2}))
		check(fn(&AssertionRange{1, 2}))
		check(fn(AssertionRange{"a", "b"}))
		check(fn(AssertionList([]interface{}{1, 2})))
	}

	t.Run("formatTypes", func(t *testing.T) {
		checkAll(t, formatTyped)
	})

	t.Run("formatValue", func(t *testing.T) {
		checkAll(t, formatValue)
	})

	t.Run("formatString", func(t *testing.T) {
		checkAll(t, formatString)
	})

	t.Run("formatRange", func(t *testing.T) {
		checkAll(t, func(v interface{}) string {
			return strings.Join(formatRange(v), "")
		})
	})

	t.Run("formatList", func(t *testing.T) {
		checkAll(t, func(v interface{}) string {
			return strings.Join(formatList(v), "")
		})
	})
}

func TestFormatDiff(t *testing.T) {
	checkOK := func(a, b interface{}) {
		s, ok := formatDiff(a, b)
		assert.True(t, ok)
		assert.NotEqual(t, "", s)
	}

	checkNotOK := func(a, b interface{}) {
		s, ok := formatDiff(a, b)
		assert.False(t, ok)
		assert.Equal(t, "", s)
	}

	checkNotOK(map[string]interface{}{}, []interface{}{})
	checkNotOK([]interface{}{}, map[string]interface{}{})
	checkNotOK("foo", "bar")
	checkNotOK(func() {}, func() {})

	checkNotOK(map[string]interface{}{}, map[string]interface{}{})
	checkNotOK([]interface{}{}, []interface{}{})

	checkOK(map[string]interface{}{"a": 1}, map[string]interface{}{})
	checkOK([]interface{}{"a"}, []interface{}{})
}

func TestFormatDataFailureActual(t *testing.T) {
	df := &DefaultFormatter{}
	ctx := &AssertionContext{}

	t.Run("assert type integer", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertType,
			Actual: &AssertionValue{
				Value: int(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveActual)
		assert.Equal(t, fd.Actual, "int(1000000)")

	})

	t.Run("assert type float32", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertType,
			Actual: &AssertionValue{
				Value: float32(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveActual)
		assert.Equal(t, fd.Actual, "float32(1e+06)")
	})

	t.Run("assert type float64", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertType,
			Actual: &AssertionValue{
				Value: float64(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveActual)
		assert.Equal(t, fd.Actual, "float64(1e+06)")
	})

	t.Run("assert type string", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertType,
			Actual: &AssertionValue{
				Value: "value string",
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveActual)
		assert.Equal(t, fd.Actual, `string("value string")`)
	})
}

func TestFormatDataFailureExpected(t *testing.T) {
	df := &DefaultFormatter{}
	ctx := &AssertionContext{}

	t.Run("assert in range integer", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertInRange,
			Expected: &AssertionValue{
				AssertionRange{
					Min: int(1_000_000),
					Max: int(2_000_000),
				},
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveExpected)
		assert.Equal(t, fd.ExpectedKind, kindRange)
		assert.Equal(t, fd.Expected, []string{"[1000000; 2000000]"})
	})

	t.Run("assert in range float32", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertInRange,
			Expected: &AssertionValue{
				AssertionRange{
					Min: float32(1_000_000),
					Max: float32(2_000_000),
				},
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveExpected)
		assert.Equal(t, fd.ExpectedKind, kindRange)
		assert.Equal(t, fd.Expected, []string{"[1e+06; 2e+06]"})
	})

	t.Run("assert in range float64", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertInRange,
			Expected: &AssertionValue{
				AssertionRange{
					Min: float64(1_000_000),
					Max: float64(2_000_000),
				},
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveExpected)
		assert.Equal(t, fd.ExpectedKind, kindRange)
		assert.Equal(t, fd.Expected, []string{"[1e+06; 2e+06]"})
	})

	t.Run("assert in range string", func(t *testing.T) {
		fl := &AssertionFailure{
			Type: AssertInRange,
			Expected: &AssertionValue{
				AssertionRange{
					Min: "min string",
					Max: "max string",
				},
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveExpected)
		assert.Equal(t, fd.ExpectedKind, kindRange)
		assert.Equal(t, fd.Expected, []string{"min string", "max string"})
	})
}

func TestFormatDataFailureDelta(t *testing.T) {
	df := &DefaultFormatter{}
	ctx := &AssertionContext{}

	t.Run("integer", func(t *testing.T) {
		fl := &AssertionFailure{
			Delta: &AssertionValue{
				Value: int(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveDelta)
		assert.Equal(t, fd.Delta, "1000000")
	})

	t.Run("float32", func(t *testing.T) {
		fl := &AssertionFailure{
			Delta: &AssertionValue{
				Value: float32(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveDelta)
		assert.Equal(t, fd.Delta, "1000000.000000")

	})

	t.Run("float64", func(t *testing.T) {
		fl := &AssertionFailure{
			Delta: &AssertionValue{
				Value: float64(1_000_000),
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveDelta)
		assert.Equal(t, fd.Delta, "1000000.000000")
	})

	t.Run("string", func(t *testing.T) {
		fl := &AssertionFailure{
			Delta: &AssertionValue{
				Value: "delta string",
			},
		}
		fd := df.buildFormatData(ctx, fl)
		assert.True(t, fd.HaveDelta)
		assert.Equal(t, fd.Delta, `"delta string"`)
	})
}
