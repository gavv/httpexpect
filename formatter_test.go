package httpexpect

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

var mockDefaultFormatter = &DefaultFormatter{}

type typedNil int

func (*typedNil) String() string {
	return ""
}

func TestFormat_Values(t *testing.T) {
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
		checkAll(t, mockDefaultFormatter.formatTyped)
	})

	t.Run("formatValue", func(t *testing.T) {
		checkAll(t, mockDefaultFormatter.formatValue)
	})

	t.Run("formatBareString", func(t *testing.T) {
		checkAll(t, mockDefaultFormatter.formatBareString)
	})

	t.Run("formatRange", func(t *testing.T) {
		checkAll(t, func(v interface{}) string {
			return strings.Join(mockDefaultFormatter.formatRange(v), "")
		})
	})

	t.Run("formatList", func(t *testing.T) {
		checkAll(t, func(v interface{}) string {
			return strings.Join(mockDefaultFormatter.formatList(v), "")
		})
	})
}

func TestFormat_Diff(t *testing.T) {
	checkOK := func(a, b interface{}) {
		s, ok := mockDefaultFormatter.formatDiff(a, b)
		assert.True(t, ok)
		assert.NotEqual(t, "", s)
	}

	checkNotOK := func(a, b interface{}) {
		s, ok := mockDefaultFormatter.formatDiff(a, b)
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

func TestFormat_FailureActual(t *testing.T) {
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
			wantActual:     "float32(1e+06)",
		},
		{
			name:           "AssertType float64",
			assertionType:  AssertType,
			assertionValue: float64(1_000_000),
			wantHaveActual: true,
			wantActual:     "float64(1e+06)",
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
			wantActual:     "1e+06",
		},
		{
			name:           "AssertValid float64",
			assertionType:  AssertValid,
			assertionValue: float64(1_000_000),
			wantHaveActual: true,
			wantActual:     "1e+06",
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

func TestFormat_FailureExpected(t *testing.T) {
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
			wantExpected:     []string{"[1e+06; 2e+06]"},
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
			wantExpected:     []string{"[1e+06; 2e+06]"},
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
			wantExpected:     []string{"1e+06"},
		},
		{
			name:             "AssertMatchPath float64",
			assertionType:    AssertMatchPath,
			assertionValue:   float64(1_000_000),
			wantHaveExpected: true,
			wantExpectedKind: kindPath,
			wantExpected:     []string{"1e+06"},
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
			wantExpected:     []string{"1e+06", "2e+06"},
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
			wantExpected:     []string{"1e+06", "2e+06"},
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

func TestFormat_FailureReference(t *testing.T) {
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
			wantReference:     "1e+06",
		},
		{
			name:              "float64",
			assertionValue:    float64(1_000_000),
			wantHaveReference: true,
			wantReference:     "1e+06",
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

func TestFormat_FailureDelta(t *testing.T) {
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
			wantDelta:      "1e+06",
		},
		{
			name:           "float64",
			assertionValue: float64(1_000_000),
			wantHaveDelta:  true,
			wantDelta:      "1e+06",
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

func TestFormatter_FormatFailure(t *testing.T) {
	tests := []struct {
		name             string
		assertionType    AssertionType
		assertionValue   interface{}
		assertionFailure AssertionFailure
		formatter        DefaultFormatter
		wantTpl          string
	}{
		{
			name: "AssertInRange float32 true",
			formatter: DefaultFormatter{
				DisableScientific: true,
			},
			assertionFailure: AssertionFailure{
				Type: AssertInRange,
				Expected: &AssertionValue{
					Value: AssertionRange{
						Min: float32(-1234567.89),
						Max: float32(1234567.89),
					},
				},
			},
			wantTpl: "[-1234567.875; 1234567.875]",
		},
		{
			name: "AssertInRange float32 false",
			formatter: DefaultFormatter{
				DisableScientific: false,
			},
			assertionFailure: AssertionFailure{
				Type: AssertInRange,
				Expected: &AssertionValue{
					Value: AssertionRange{
						Min: float32(-1234567.89),
						Max: float32(1234567.89),
					},
				},
			},
			wantTpl: "[-1.2345679e+06; 1.2345679e+06]",
		},
		{
			name: "AssertInRange float64 true",
			formatter: DefaultFormatter{
				DisableScientific: true,
			},
			assertionFailure: AssertionFailure{
				Type: AssertInRange,
				Expected: &AssertionValue{
					Value: AssertionRange{
						Min: float64(-1234567.89),
						Max: float64(1234567.89),
					},
				},
			},
			wantTpl: "[-1234567.89; 1234567.89]",
		},
		{
			name: "AssertInRange float64 false",
			formatter: DefaultFormatter{
				DisableScientific: false,
			},
			assertionFailure: AssertionFailure{
				Type: AssertInRange,
				Expected: &AssertionValue{
					Value: AssertionRange{
						Min: float64(-1234567.89),
						Max: float64(1234567.89),
					},
				},
			},
			wantTpl: "[-1.23456789e+06; 1.23456789e+06]",
		},
	}

	ctx := &AssertionContext{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			tpl := tc.formatter.FormatFailure(ctx, &tc.assertionFailure)
			assert.Contains(t, tpl, tc.wantTpl)
		})
	}
}
