package httpexpect

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type typedStingerNil int

func (*typedStingerNil) String() string {
	return ""
}

func TestFormatter_FailureActual(t *testing.T) {
	tests := []struct {
		name           string
		assertionType  AssertionType
		assertionValue interface{}
		wantHaveActual bool
		wantActual     string
	}{
		// AssertType
		{
			name:           "AssertType nil",
			assertionType:  AssertType,
			assertionValue: nil,
			wantHaveActual: true,
			wantActual:     "<nil>(<nil>)",
		},
		{
			name:           "AssertType int",
			assertionType:  AssertType,
			assertionValue: int(1_000_000),
			wantHaveActual: true,
			wantActual:     "int(1000000)",
		},
		{
			name:           "AssertType float32",
			assertionType:  AssertType,
			assertionValue: float32(1_000_000),
			wantHaveActual: true,
			wantActual:     "float32(1000000)",
		},
		{
			name:           "AssertType float64",
			assertionType:  AssertType,
			assertionValue: float64(1_000_000),
			wantHaveActual: true,
			wantActual:     "float64(1000000)",
		},
		{
			name:           "AssertType string",
			assertionType:  AssertType,
			assertionValue: "test string",
			wantHaveActual: true,
			wantActual:     "string(\"test string\")",
		},
		{
			name:           "AssertType object",
			assertionType:  AssertType,
			assertionValue: struct{ Name string }{"test name"},
			wantHaveActual: true,
			wantActual:     "struct { Name string }(struct { Name string }{Name:\"test name\"})",
		},

		// AssertValid
		{
			name:           "AssertValid nil",
			assertionType:  AssertValid,
			assertionValue: nil,
			wantHaveActual: true,
			wantActual:     "nil",
		},
		{
			name:           "AssertValid int",
			assertionType:  AssertValid,
			assertionValue: int(1_000_000),
			wantHaveActual: true,
			wantActual:     "1000000",
		},
		{
			name:           "AssertValid float32",
			assertionType:  AssertValid,
			assertionValue: float32(1_000_000),
			wantHaveActual: true,
			wantActual:     "1000000",
		},
		{
			name:           "AssertValid float64",
			assertionType:  AssertValid,
			assertionValue: float64(1_000_000),
			wantHaveActual: true,
			wantActual:     "1000000",
		},
		{
			name:           "AssertValid string",
			assertionType:  AssertValid,
			assertionValue: "test string",
			wantHaveActual: true,
			wantActual:     "\"test string\"",
		},
		{
			name:           "AssertValid object",
			assertionType:  AssertValid,
			assertionValue: struct{ Name string }{"test name"},
			wantHaveActual: true,
			wantActual:     "{\n  \"Name\": \"test name\"\n}",
		},
	}

	df := &DefaultFormatter{}
	ctx := &AssertionContext{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fl := &AssertionFailure{
				Type: tc.assertionType,
				Actual: &AssertionValue{
					Value: tc.assertionValue,
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.Equal(t, tc.wantHaveActual, fd.HaveActual)
			assert.Equal(t, tc.wantActual, fd.Actual)
		})
	}
}

func TestFormatter_FailureExpected(t *testing.T) {
	tests := []struct {
		name             string
		assertionType    AssertionType
		assertionValue   interface{}
		formatter        DefaultFormatter
		wantHaveExpected bool
		wantExpectedKind string
		wantExpected     []string
	}{
		// AssertInRange
		{
			name:          "AssertInRange nil",
			assertionType: AssertInRange,
			assertionValue: AssertionRange{
				Min: nil,
				Max: nil,
			},
			wantHaveExpected: true,
			wantExpectedKind: kindRange,
			wantExpected:     []string{"<nil>", "<nil>"},
		},
		{
			name:          "AssertInRange int",
			assertionType: AssertInRange,
			assertionValue: AssertionRange{
				Min: int(1_000_000),
				Max: int(2_000_000),
			},
			wantHaveExpected: true,
			wantExpectedKind: kindRange,
			wantExpected:     []string{"[1000000; 2000000]"},
		},
		{
			name:          "AssertInRange float32",
			assertionType: AssertInRange,
			assertionValue: AssertionRange{
				Min: float32(1_000_000),
				Max: float32(2_000_000),
			},
			wantHaveExpected: true,
			wantExpectedKind: kindRange,
			wantExpected:     []string{"[1000000; 2000000]"},
		},
		{
			name:          "AssertInRange float64",
			assertionType: AssertInRange,
			assertionValue: AssertionRange{
				Min: float64(1_000_000),
				Max: float64(2_000_000),
			},
			wantHaveExpected: true,
			wantExpectedKind: kindRange,
			wantExpected:     []string{"[1000000; 2000000]"},
		},
		{
			name:          "AssertInRange string",
			assertionType: AssertInRange,
			assertionValue: AssertionRange{
				Min: "test string 1",
				Max: "test string 2",
			},
			wantHaveExpected: true,
			wantExpectedKind: kindRange,
			wantExpected:     []string{"test string 1", "test string 2"},
		},
		{
			name:          "AssertInRange object",
			assertionType: AssertInRange,
			assertionValue: AssertionRange{
				Min: struct{ Name string }{"test name 1"},
				Max: struct{ Name string }{"test name 2"},
			},
			wantHaveExpected: true,
			wantExpectedKind: kindRange,
			wantExpected:     []string{"{test name 1}", "{test name 2}"},
		},

		// AssertMatchPath
		{
			name:             "AssertMatchPath nil",
			assertionType:    AssertMatchPath,
			assertionValue:   nil,
			wantHaveExpected: true,
			wantExpectedKind: kindPath,
			wantExpected:     []string{"nil"},
		},
		{
			name:             "AssertMatchPath int",
			assertionType:    AssertMatchPath,
			assertionValue:   int(1_000_000),
			wantHaveExpected: true,
			wantExpectedKind: kindPath,
			wantExpected:     []string{"1000000"},
		},
		{
			name:             "AssertMatchPath float32",
			assertionType:    AssertMatchPath,
			assertionValue:   float32(1_000_000),
			wantHaveExpected: true,
			wantExpectedKind: kindPath,
			wantExpected:     []string{"1000000"},
		},
		{
			name:             "AssertMatchPath float64",
			assertionType:    AssertMatchPath,
			assertionValue:   float64(1_000_000),
			wantHaveExpected: true,
			wantExpectedKind: kindPath,
			wantExpected:     []string{"1000000"},
		},
		{
			name:             "AssertMatchPath string",
			assertionType:    AssertMatchPath,
			assertionValue:   "test string",
			wantHaveExpected: true,
			wantExpectedKind: kindPath,
			wantExpected:     []string{"test string"},
		},
		{
			name:             "AssertMatchPath object",
			assertionType:    AssertMatchPath,
			assertionValue:   struct{ Name string }{"test name"},
			wantHaveExpected: true,
			wantExpectedKind: kindPath,
			wantExpected:     []string{"{\n  \"Name\": \"test name\"\n}"},
		},

		// AssertMatchFormat
		{
			name:          "AssertMatchFormat nil",
			assertionType: AssertMatchFormat,
			assertionValue: AssertionList{
				nil,
				nil,
			},
			wantHaveExpected: true,
			wantExpectedKind: kindFormatList,
			wantExpected:     []string{"nil", "nil"},
		},
		{
			name:          "AssertMatchFormat int",
			assertionType: AssertMatchFormat,
			assertionValue: AssertionList{
				int(1_000_000),
				int(2_000_000),
			},
			wantHaveExpected: true,
			wantExpectedKind: kindFormatList,
			wantExpected:     []string{"1000000", "2000000"},
		},
		{
			name:          "AssertMatchFormat float32",
			assertionType: AssertMatchFormat,
			assertionValue: AssertionList{
				float32(1_000_000),
				float32(2_000_000),
			},
			wantHaveExpected: true,
			wantExpectedKind: kindFormatList,
			wantExpected:     []string{"1000000", "2000000"},
		},
		{
			name:          "AssertMatchFormat float64",
			assertionType: AssertMatchFormat,
			assertionValue: AssertionList{
				float64(1_000_000),
				float64(2_000_000),
			},
			wantHaveExpected: true,
			wantExpectedKind: kindFormatList,
			wantExpected:     []string{"1000000", "2000000"},
		},
		{
			name:          "AssertMatchFormat string",
			assertionType: AssertMatchFormat,
			assertionValue: AssertionList{
				"test string 1",
				"test string 2",
			},
			wantHaveExpected: true,
			wantExpectedKind: kindFormatList,
			wantExpected:     []string{"\"test string 1\"", "\"test string 2\""},
		},
		{
			name:          "AssertMatchFormat object",
			assertionType: AssertMatchFormat,
			assertionValue: AssertionList{
				struct{ Name string }{"test name 1"},
				struct{ Name string }{"test name 2"},
			},
			wantHaveExpected: true,
			wantExpectedKind: kindFormatList,
			wantExpected: []string{
				"{\n  \"Name\": \"test name 1\"\n}", "{\n  \"Name\": \"test name 2\"\n}",
			},
		},
	}

	ctx := &AssertionContext{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fl := &AssertionFailure{
				Type: tc.assertionType,
				Expected: &AssertionValue{
					Value: tc.assertionValue,
				},
			}
			fd := tc.formatter.buildFormatData(ctx, fl)
			assert.Equal(t, tc.wantHaveExpected, fd.HaveExpected)
			assert.Equal(t, tc.wantExpectedKind, fd.ExpectedKind)
			assert.Equal(t, tc.wantExpected, fd.Expected)
		})
	}
}

func TestFormatter_FailureReference(t *testing.T) {
	tests := []struct {
		name              string
		assertionValue    interface{}
		wantHaveReference bool
		wantReference     string
	}{
		{
			name:              "nil",
			assertionValue:    nil,
			wantHaveReference: true,
			wantReference:     "nil",
		},
		{
			name:              "int",
			assertionValue:    int(1_000_000),
			wantHaveReference: true,
			wantReference:     "1000000",
		},
		{
			name:              "float32",
			assertionValue:    float32(1_000_000),
			wantHaveReference: true,
			wantReference:     "1000000",
		},
		{
			name:              "float64",
			assertionValue:    float64(1_000_000),
			wantHaveReference: true,
			wantReference:     "1000000",
		},
		{
			name:              "string",
			assertionValue:    "test string",
			wantHaveReference: true,
			wantReference:     "\"test string\"",
		},
		{
			name:              "object",
			assertionValue:    struct{ Name string }{"test name"},
			wantHaveReference: true,
			wantReference:     "{\n  \"Name\": \"test name\"\n}",
		},
	}

	df := &DefaultFormatter{}
	ctx := &AssertionContext{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fl := &AssertionFailure{
				Reference: &AssertionValue{
					Value: tc.assertionValue,
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.Equal(t, tc.wantHaveReference, fd.HaveReference)
			assert.Equal(t, tc.wantReference, fd.Reference)
		})
	}
}

func TestFormatter_FailureDelta(t *testing.T) {
	tests := []struct {
		name           string
		assertionValue interface{}
		wantHaveDelta  bool
		wantDelta      string
	}{
		{
			name:           "nil",
			assertionValue: nil,
			wantHaveDelta:  true,
			wantDelta:      "nil",
		},
		{
			name:           "int",
			assertionValue: int(1_000_000),
			wantHaveDelta:  true,
			wantDelta:      "1000000",
		},
		{
			name:           "float32",
			assertionValue: float32(1_000_000),
			wantHaveDelta:  true,
			wantDelta:      "1000000",
		},
		{
			name:           "float64",
			assertionValue: float64(1_000_000),
			wantHaveDelta:  true,
			wantDelta:      "1000000",
		},
		{
			name:           "string",
			assertionValue: "test string",
			wantHaveDelta:  true,
			wantDelta:      "\"test string\"",
		},
		{
			name:           "object",
			assertionValue: struct{ Name string }{"test name"},
			wantHaveDelta:  true,
			wantDelta:      "{\n  \"Name\": \"test name\"\n}",
		},
	}

	df := &DefaultFormatter{}
	ctx := &AssertionContext{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fl := &AssertionFailure{
				Delta: &AssertionValue{
					Value: tc.assertionValue,
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.Equal(t, tc.wantHaveDelta, fd.HaveDelta)
			assert.Equal(t, tc.wantDelta, fd.Delta)
		})
	}
}

func TestFormatter_FloatFormat(t *testing.T) {
	type testCase struct {
		name     string
		format   FloatFormat
		value    interface{}
		wantText string
	}

	testCases := []testCase{
		// float32
		{
			name:     "float32 auto small exponent",
			format:   FloatFormatAuto,
			value:    float32(1.2345678),
			wantText: "1.2345678",
		},
		{
			name:     "float32 auto large exponent",
			format:   FloatFormatAuto,
			value:    float32(1234567.8),
			wantText: "1.2345678e+06",
		},
		{
			name:     "float32 decimal",
			format:   FloatFormatDecimal,
			value:    float32(1234567.8),
			wantText: "1234567.8",
		},
		{
			name:     "float32 scientific",
			format:   FloatFormatScientific,
			value:    float32(1.2345678),
			wantText: "1.2345678e+00",
		},
		// float64
		{
			name:     "float64 auto small exponent",
			format:   FloatFormatAuto,
			value:    float64(1.23456789),
			wantText: "1.23456789",
		},
		{
			name:     "float64 auto large exponent",
			format:   FloatFormatAuto,
			value:    float64(12345678.9),
			wantText: "1.23456789e+07",
		},
		{
			name:     "float64 decimal",
			format:   FloatFormatDecimal,
			value:    float64(12345678.9),
			wantText: "12345678.9",
		},
		{
			name:     "float64 scientific",
			format:   FloatFormatScientific,
			value:    float64(1.23456789),
			wantText: "1.23456789e+00",
		},
		// no fractional part
		{
			name:     "nofrac auto",
			format:   FloatFormatAuto,
			value:    float32(12345678),
			wantText: "12345678",
		},
		{
			name:     "nofrac decimal",
			format:   FloatFormatDecimal,
			value:    float32(12345678),
			wantText: "12345678",
		},
		{
			name:     "nofrac scientific",
			format:   FloatFormatScientific,
			value:    float32(12345678),
			wantText: "1.2345678e+07",
		},
		// integer
		{
			name:     "integer auto",
			format:   FloatFormatAuto,
			value:    int(12345678),
			wantText: "12345678",
		},
		{
			name:     "integer decimal",
			format:   FloatFormatDecimal,
			value:    int(12345678),
			wantText: "12345678",
		},
		{
			name:     "integer scientific",
			format:   FloatFormatScientific,
			value:    int(12345678),
			wantText: "12345678",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := DefaultFormatter{
				FloatFormat: tc.format,
			}
			formatData := formatter.buildFormatData(
				&AssertionContext{},
				&AssertionFailure{
					Type:   AssertValid,
					Actual: &AssertionValue{tc.value},
				})
			assert.Equal(t, tc.wantText, formatData.Actual)
		})
	}
}

func TestFormatter_FloatFields(t *testing.T) {
	type testCase struct {
		name     string
		format   FloatFormat
		value    float64
		wantText string
	}

	testCases := []testCase{
		{
			name:     "auto",
			format:   FloatFormatAuto,
			value:    1.2345678,
			wantText: "1.2345678",
		},
		{
			name:     "decimal",
			format:   FloatFormatDecimal,
			value:    1.2345678,
			wantText: "1.2345678",
		},
		{
			name:     "scientific",
			format:   FloatFormatScientific,
			value:    1.2345678,
			wantText: "1.2345678e+00",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := DefaultFormatter{
				FloatFormat: tc.format,
			}

			t.Run("actual", func(t *testing.T) {
				formatData := formatter.buildFormatData(
					&AssertionContext{},
					&AssertionFailure{
						Type:   AssertValid,
						Actual: &AssertionValue{tc.value},
					})
				assert.Equal(t, tc.wantText, formatData.Actual)
			})

			t.Run("expected", func(t *testing.T) {
				formatData := formatter.buildFormatData(
					&AssertionContext{},
					&AssertionFailure{
						Type:     AssertEqual,
						Actual:   &AssertionValue{},
						Expected: &AssertionValue{tc.value},
					})
				require.Equal(t, 1, len(formatData.Expected))
				assert.Equal(t, tc.wantText, formatData.Expected[0])
			})

			t.Run("reference", func(t *testing.T) {
				formatData := formatter.buildFormatData(
					&AssertionContext{},
					&AssertionFailure{
						Type:      AssertEqual,
						Actual:    &AssertionValue{},
						Expected:  &AssertionValue{},
						Reference: &AssertionValue{tc.value},
					})
				assert.Equal(t, tc.wantText, formatData.Reference)
			})

			t.Run("delta", func(t *testing.T) {
				formatData := formatter.buildFormatData(
					&AssertionContext{},
					&AssertionFailure{
						Type:     AssertEqual,
						Actual:   &AssertionValue{},
						Expected: &AssertionValue{},
						Delta:    &AssertionValue{tc.value},
					})
				assert.Equal(t, tc.wantText, formatData.Delta)
			})

			t.Run("range", func(t *testing.T) {
				formatData := formatter.buildFormatData(
					&AssertionContext{},
					&AssertionFailure{
						Type:   AssertInRange,
						Actual: &AssertionValue{},
						Expected: &AssertionValue{AssertionRange{
							Min: tc.value,
							Max: tc.value,
						}},
					})
				require.Equal(t, 1, len(formatData.Expected))
				assert.Equal(t,
					fmt.Sprintf("[%s; %s]", tc.wantText, tc.wantText),
					formatData.Expected[0])
			})

			t.Run("list", func(t *testing.T) {
				formatData := formatter.buildFormatData(
					&AssertionContext{},
					&AssertionFailure{
						Type:   AssertBelongs,
						Actual: &AssertionValue{},
						Expected: &AssertionValue{AssertionList{
							tc.value,
							tc.value,
						}},
					})
				require.Equal(t, 2, len(formatData.Expected))
				assert.Equal(t, tc.wantText, formatData.Expected[0])
				assert.Equal(t, tc.wantText, formatData.Expected[1])
			})
		})
	}
}

func TestFormatter_FormatValue(t *testing.T) {
	var formatter = &DefaultFormatter{}

	checkAll := func(t *testing.T, fn func(interface{}) string) {
		var tnil *typedStingerNil
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
		check(fn(float32(123)))
		check(fn(float64(123)))
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

	t.Run("formatValue", func(t *testing.T) {
		checkAll(t, formatter.formatValue)
	})

	t.Run("formatTypedValue", func(t *testing.T) {
		checkAll(t, formatter.formatTypedValue)
	})

	t.Run("formatMatchValue", func(t *testing.T) {
		checkAll(t, formatter.formatMatchValue)
	})

	t.Run("formatRangeValue", func(t *testing.T) {
		checkAll(t, func(v interface{}) string {
			return strings.Join(formatter.formatRangeValue(v), "")
		})
	})

	t.Run("formatListValue", func(t *testing.T) {
		checkAll(t, func(v interface{}) string {
			return strings.Join(formatter.formatListValue(v), "")
		})
	})
}

func TestFormatter_FormatDiff(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		check := func(a, b interface{}) {
			formatter := &DefaultFormatter{}
			diff, ok := formatter.formatDiff(a, b)
			assert.True(t, ok)
			assert.NotEqual(t, "", diff)
		}

		check(map[string]interface{}{"a": 1}, map[string]interface{}{})
		check([]interface{}{"a"}, []interface{}{})
	})

	t.Run("failure", func(t *testing.T) {
		check := func(a, b interface{}) {
			formatter := &DefaultFormatter{}
			diff, ok := formatter.formatDiff(a, b)
			assert.False(t, ok)
			assert.Equal(t, "", diff)
		}

		check(map[string]interface{}{}, []interface{}{})
		check([]interface{}{}, map[string]interface{}{})
		check("foo", "bar")
		check(func() {}, func() {})

		check(map[string]interface{}{}, map[string]interface{}{})
		check([]interface{}{}, []interface{}{})
	})
}

func TestFormatter_ColorMode(t *testing.T) {
	cases := []struct {
		name string
		mode ColorMode
		want bool
	}{
		{
			name: "always",
			mode: ColorModeAlways,
			want: true,
		},
		{
			name: "never",
			mode: ColorModeNever,
			want: false,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := DefaultFormatter{
				ColorMode: tc.mode,
			}
			fd := f.buildFormatData(&AssertionContext{}, &AssertionFailure{})
			assert.Equal(t, tc.want, fd.EnableColors)
		})
	}
}
