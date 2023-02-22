package httpexpect

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNumber_FailedChain(t *testing.T) {
	chain := newMockChain(t)
	chain.setFailed()

	value := newNumber(chain, 0)
	value.chain.assertFailed(t)

	value.Path("$").chain.assertFailed(t)
	value.Schema("")
	value.Alias("foo")

	var target interface{}
	value.Decode(&target)

	value.IsEqual(0)
	value.NotEqual(0)
	value.InDelta(0, 0)
	value.NotInDelta(0, 0)
	value.InRange(0, 0)
	value.NotInRange(0, 0)
	value.InList(0)
	value.NotInList(0)
	value.Gt(0)
	value.Ge(0)
	value.Lt(0)
	value.Le(0)
	value.IsInt()
	value.NotInt()
	value.IsUint()
	value.NotUint()
	value.IsFinite()
	value.NotFinite()
}

func TestNumber_Constructors(t *testing.T) {
	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewNumber(reporter, 10.3)
		value.IsEqual(10.3)
		value.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewNumberC(Config{
			Reporter: reporter,
		}, 10.3)
		value.IsEqual(10.3)
		value.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newNumber(chain, 10.3)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})
}

func TestNumber_Decode(t *testing.T) {
	t.Run("target is empty interface", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10.1)

		var target interface{}
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, 10.1, target)
	})

	t.Run("target is int", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10)

		var target int
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, 10, target)
	})

	t.Run("target is float64", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10.1)

		var target float64
		value.Decode(&target)

		value.chain.assertNotFailed(t)
		assert.Equal(t, 10.1, target)
	})

	t.Run("target is nil", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10.1)

		value.Decode(nil)

		value.chain.assertFailed(t)
	})

	t.Run("target is unmarshable", func(t *testing.T) {
		reporter := newMockReporter(t)

		value := NewNumber(reporter, 10.1)

		value.Decode(123)

		value.chain.assertFailed(t)
	})
}

func TestNumber_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 123)
	assert.Equal(t, []string{"Number()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Number()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Number()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)
}

func TestNumber_Getters(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 123.0)

	assert.Equal(t, 123.0, value.Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	assert.Equal(t, 123.0, value.Path("$").Raw())
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Schema(`{"type": "number"}`)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Schema(`{"type": "object"}`)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestNumber_IsEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	assert.Equal(t, 1234, int(value.Raw()))

	value.IsEqual(1234)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsEqual(4321)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(4321)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(1234)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestNumber_EqualNaN(t *testing.T) {
	reporter := newMockReporter(t)

	v1 := NewNumber(reporter, math.NaN())
	v1.IsEqual(1234.5)
	v1.chain.assertFailed(t)

	v2 := NewNumber(reporter, 1234.5)
	v2.IsEqual(math.NaN())
	v2.chain.assertFailed(t)

	v3 := NewNumber(reporter, math.NaN())
	v3.InDelta(1234.0, 0.1)
	v3.chain.assertFailed(t)

	v4 := NewNumber(reporter, 1234.5)
	v4.InDelta(math.NaN(), 0.1)
	v4.chain.assertFailed(t)

	v5 := NewNumber(reporter, 1234.5)
	v5.InDelta(1234.5, math.NaN())
	v5.chain.assertFailed(t)

	v6 := NewNumber(reporter, math.NaN())
	v6.NotInDelta(1234.0, 0.1)
	v6.chain.assertFailed(t)

	v7 := NewNumber(reporter, 1234.5)
	v7.NotInDelta(math.NaN(), 0.1)
	v7.chain.assertFailed(t)

	v8 := NewNumber(reporter, 1234.5)
	v8.NotInDelta(1234.5, math.NaN())
	v8.chain.assertFailed(t)
}

func TestNumber_InDelta(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234.5)

	value.InDelta(1234.3, 0.3)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InDelta(1234.7, 0.3)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InDelta(1234.3, 0.1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InDelta(1234.7, 0.1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInDelta(1234.3, 0.3)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInDelta(1234.7, 0.3)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInDelta(1234.3, 0.1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInDelta(1234.7, 0.1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestNumber_InRange(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.InRange(1234, 1234)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(1234, 1234)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(1234-1, 1234)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(1234-1, 1234)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(1234, 1234+1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(1234, 1234+1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(1234+1, 1234+2)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(1234+1, 1234+2)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InRange(1234-2, 1234-1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(1234-2, 1234-1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InRange(1234+1, 1234-1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(1234+1, 1234-1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(1234+1, "NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange("NOT NUMBER", 1234+2)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestNumber_InList(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.InList()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(1234, 4567)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInList(1234, 4567)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(1234.00, 4567.00)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInList(1234.00, 4567.00)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(4567.00, 1234.01)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList(4567.00, 1234.01)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InList(1234+1, "NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList("NOT NUMBER", 1234+2)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestNumber_IsGreater(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.Gt(1234 - 1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Gt(1234)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Ge(1234 - 1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Ge(1234)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Ge(1234 + 1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestNumber_IsLesser(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.Lt(1234 + 1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Lt(1234)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Le(1234 + 1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Le(1234)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Le(1234 - 1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestNumber_ConvertEqual(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.IsEqual(int64(1234))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsEqual(float32(1234))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsEqual("NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(int64(4321))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(float32(4321))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual("NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestNumber_ConvertInRange(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.InRange(int64(1233), float32(1235))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(int64(1233), float32(1235))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(1235, 1236)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(1235, 1236)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InRange(int64(1233), "NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(int64(1233), "NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange("NOT NUMBER", float32(1235))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange("NOT NUMBER", float32(1235))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

}

func TestNumber_ConvertInList(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 111)

	value.InList(int64(111), float32(222))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInList(int64(111), float32(222))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(float32(111), int64(222))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInList(float32(111), int64(222))
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(222, 333)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList(222, 333)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InList(222, "NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList(222, "NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(111, "NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList(111, "NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestNumber_ConvertGreater(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.Gt(int64(1233))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Gt(float32(1233))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Gt("NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Ge(int64(1233))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Ge(float32(1233))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Ge("NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestNumber_ConvertLesser(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewNumber(reporter, 1234)

	value.Lt(int64(1235))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Lt(float32(1235))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Lt("NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Le(int64(1235))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Le(float32(1235))
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Le("NOT NUMBER")
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestNumber_IsInt(t *testing.T) {
	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).IsInt(16, 32).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotInt(16, 32).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).IsInt(0).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotInt(0).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).IsInt(-16).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotInt(-16).
			chain.assertFailed(t)
	})

	t.Run("values", func(t *testing.T) {
		tests := []struct {
			name    string
			value   float64
			isInt16 bool
			isInt32 bool
			isInt   bool
		}{
			{
				name:    "0",
				value:   0,
				isInt16: true,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "1",
				value:   1,
				isInt16: true,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "0.5",
				value:   0.5,
				isInt16: false,
				isInt32: false,
				isInt:   false,
			},
			{
				name:    "NaN",
				value:   math.NaN(),
				isInt16: false,
				isInt32: false,
				isInt:   false,
			},
			{
				name:    "-Inf",
				value:   math.Inf(-1),
				isInt16: false,
				isInt32: false,
				isInt:   false,
			},
			{
				name:    "+Inf",
				value:   math.Inf(+1),
				isInt16: false,
				isInt32: false,
				isInt:   false,
			},
			{
				name:    "MinInt16-1",
				value:   math.MinInt16 - 1,
				isInt16: false,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "MinInt16",
				value:   math.MinInt16,
				isInt16: true,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "MaxInt16",
				value:   math.MaxInt16,
				isInt16: true,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "MaxInt16+1",
				value:   math.MaxInt16 + 1,
				isInt16: false,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "MinInt32-1",
				value:   math.MinInt32 - 1,
				isInt16: false,
				isInt32: false,
				isInt:   true,
			},
			{
				name:    "MinInt32",
				value:   math.MinInt32,
				isInt16: false,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "MaxInt32",
				value:   math.MaxInt32,
				isInt16: false,
				isInt32: true,
				isInt:   true,
			},
			{
				name:    "MaxInt32+1",
				value:   math.MaxInt32 + 1,
				isInt16: false,
				isInt32: false,
				isInt:   true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				if tc.isInt {
					NewNumber(reporter, tc.value).IsInt().chain.assertNotFailed(t)
					NewNumber(reporter, tc.value).NotInt().chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.value).IsInt().chain.assertFailed(t)
					NewNumber(reporter, tc.value).NotInt().chain.assertNotFailed(t)
				}

				if tc.isInt32 {
					NewNumber(reporter, tc.value).IsInt(32).chain.assertNotFailed(t)
					NewNumber(reporter, tc.value).NotInt(32).chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.value).IsInt(32).chain.assertFailed(t)
					NewNumber(reporter, tc.value).NotInt(32).chain.assertNotFailed(t)
				}

				if tc.isInt16 {
					NewNumber(reporter, tc.value).IsInt(16).chain.assertNotFailed(t)
					NewNumber(reporter, tc.value).NotInt(16).chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.value).IsInt(16).chain.assertFailed(t)
					NewNumber(reporter, tc.value).NotInt(16).chain.assertNotFailed(t)
				}
			})
		}
	})
}

func TestNumber_IsUint(t *testing.T) {
	t.Run("invalid argument", func(t *testing.T) {
		reporter := newMockReporter(t)

		NewNumber(reporter, 1234).IsUint(16, 32).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotUint(16, 32).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).IsUint(0).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotUint(0).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).IsUint(-16).
			chain.assertFailed(t)

		NewNumber(reporter, 1234).NotUint(-16).
			chain.assertFailed(t)
	})

	t.Run("values", func(t *testing.T) {
		tests := []struct {
			name     string
			value    float64
			isUint16 bool
			isUint32 bool
			isUint   bool
		}{
			{
				name:     "0",
				value:    0,
				isUint16: true,
				isUint32: true,
				isUint:   true,
			},
			{
				name:     "1",
				value:    1,
				isUint16: true,
				isUint32: true,
				isUint:   true,
			},
			{
				name:     "-1",
				value:    -1,
				isUint16: false,
				isUint32: false,
				isUint:   false,
			},
			{
				name:     "0.5",
				value:    0.5,
				isUint16: false,
				isUint32: false,
				isUint:   false,
			},
			{
				name:     "NaN",
				value:    math.NaN(),
				isUint16: false,
				isUint32: false,
				isUint:   false,
			},
			{
				name:     "-Inf",
				value:    math.Inf(-1),
				isUint16: false,
				isUint32: false,
				isUint:   false,
			},
			{
				name:     "+Inf",
				value:    math.Inf(+1),
				isUint16: false,
				isUint32: false,
				isUint:   false,
			},
			{
				name:     "MaxUint16",
				value:    math.MaxUint16,
				isUint16: true,
				isUint32: true,
				isUint:   true,
			},
			{
				name:     "MaxUint16+1",
				value:    math.MaxUint16 + 1,
				isUint16: false,
				isUint32: true,
				isUint:   true,
			},
			{
				name:     "MaxUint32",
				value:    math.MaxUint32,
				isUint16: false,
				isUint32: true,
				isUint:   true,
			},
			{
				name:     "MaxUint32+1",
				value:    math.MaxUint32 + 1,
				isUint16: false,
				isUint32: false,
				isUint:   true,
			},
		}

		for _, tc := range tests {
			t.Run(tc.name, func(t *testing.T) {
				reporter := newMockReporter(t)

				if tc.isUint {
					NewNumber(reporter, tc.value).IsUint().chain.assertNotFailed(t)
					NewNumber(reporter, tc.value).NotUint().chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.value).IsUint().chain.assertFailed(t)
					NewNumber(reporter, tc.value).NotUint().chain.assertNotFailed(t)
				}

				if tc.isUint32 {
					NewNumber(reporter, tc.value).IsUint(32).chain.assertNotFailed(t)
					NewNumber(reporter, tc.value).NotUint(32).chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.value).IsUint(32).chain.assertFailed(t)
					NewNumber(reporter, tc.value).NotUint(32).chain.assertNotFailed(t)
				}

				if tc.isUint16 {
					NewNumber(reporter, tc.value).IsUint(16).chain.assertNotFailed(t)
					NewNumber(reporter, tc.value).NotUint(16).chain.assertFailed(t)
				} else {
					NewNumber(reporter, tc.value).IsUint(16).chain.assertFailed(t)
					NewNumber(reporter, tc.value).NotUint(16).chain.assertNotFailed(t)
				}
			})
		}
	})
}

func TestNumber_IsFinite(t *testing.T) {
	tests := []struct {
		name     string
		value    float64
		isFinite bool
	}{
		{
			name:     "0",
			value:    0,
			isFinite: true,
		},
		{
			name:     "1",
			value:    1,
			isFinite: true,
		},
		{
			name:     "-1",
			value:    -1,
			isFinite: true,
		},
		{
			name:     "0.5",
			value:    0.5,
			isFinite: true,
		},
		{
			name:     "NaN",
			value:    math.NaN(),
			isFinite: false,
		},
		{
			name:     "-Inf",
			value:    math.Inf(-1),
			isFinite: false,
		},
		{
			name:     "+Inf",
			value:    math.Inf(+1),
			isFinite: false,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			reporter := newMockReporter(t)

			if tc.isFinite {
				NewNumber(reporter, tc.value).IsFinite().chain.assertNotFailed(t)
				NewNumber(reporter, tc.value).NotFinite().chain.assertFailed(t)
			} else {
				NewNumber(reporter, tc.value).IsFinite().chain.assertFailed(t)
				NewNumber(reporter, tc.value).NotFinite().chain.assertNotFailed(t)
			}
		})
	}
}
