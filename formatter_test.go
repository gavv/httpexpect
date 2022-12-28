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

	t.Run("formatBareString", func(t *testing.T) {
		checkAll(t, formatBareString)
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
	tests := []struct {
		name           string
		assertionType  AssertionType
		assertionValue interface{}
		wantHaveActual bool
		wantActual     string
	}{
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
			assertionValue: "value string",
			wantHaveActual: true,
			wantActual:     "string(\"value string\")",
		},
		{
			name:           "AssertType object",
			assertionType:  AssertType,
			assertionValue: struct{ Name string }{"testName"},
			wantHaveActual: true,
			wantActual:     "struct { Name string }(struct { Name string }{Name:\"testName\"})",
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

func TestFormatDataFailureExpected(t *testing.T) {
	tests := []struct {
		name             string
		assertionType    AssertionType
		assertionValue   interface{}
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
				Min: "string 1",
				Max: "string 2",
			},
			wantHaveExpected: true,
			wantExpectedKind: kindRange,
			wantExpected:     []string{"string 1", "string 2"},
		},
		{
			name:          "AssertInRange object",
			assertionType: AssertInRange,
			assertionValue: AssertionRange{
				Min: struct{ Name string }{"testName1"},
				Max: struct{ Name string }{"testName2"},
			},
			wantHaveExpected: true,
			wantExpectedKind: kindRange,
			wantExpected:     []string{"{testName1}", "{testName2}"},
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
			assertionValue:   "match path string",
			wantHaveExpected: true,
			wantExpectedKind: kindPath,
			wantExpected:     []string{"match path string"},
		},
		{
			name:             "AssertMatchPath object",
			assertionType:    AssertMatchPath,
			assertionValue:   struct{ Name string }{"testName"},
			wantHaveExpected: true,
			wantExpectedKind: kindPath,
			wantExpected:     []string{"{\n  \"Name\": \"testName\"\n}"},
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
				"string 1",
				"string 2",
			},
			wantHaveExpected: true,
			wantExpectedKind: kindFormatList,
			wantExpected:     []string{"\"string 1\"", "\"string 2\""},
		},
		{
			name:          "AssertMatchFormat object",
			assertionType: AssertMatchFormat,
			assertionValue: AssertionList{
				struct{ Name string }{"testName1"},
				struct{ Name string }{"testName2"},
			},
			wantHaveExpected: true,
			wantExpectedKind: kindFormatList,
			wantExpected: []string{
				"{\n  \"Name\": \"testName1\"\n}", "{\n  \"Name\": \"testName2\"\n}",
			},
		},
	}

	df := &DefaultFormatter{}
	ctx := &AssertionContext{}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fl := &AssertionFailure{
				Type: tc.assertionType,
				Expected: &AssertionValue{
					Value: tc.assertionValue,
				},
			}
			fd := df.buildFormatData(ctx, fl)
			assert.Equal(t, tc.wantHaveExpected, fd.HaveExpected)
			assert.Equal(t, tc.wantExpectedKind, fd.ExpectedKind)
			assert.Equal(t, tc.wantExpected, fd.Expected)
		})
	}
}

func TestFormatDataFailureReference(t *testing.T) {
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
			assertionValue:    "reference string",
			wantHaveReference: true,
			wantReference:     "\"reference string\"",
		},
		{
			name:              "object",
			assertionValue:    struct{ Name string }{"testName"},
			wantHaveReference: true,
			wantReference:     "{\n  \"Name\": \"testName\"\n}",
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

func TestFormatDataFailureDelta(t *testing.T) {
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
			assertionValue: "delta string",
			wantHaveDelta:  true,
			wantDelta:      "\"delta string\"",
		},
		{
			name:           "object",
			assertionValue: struct{ Name string }{"testName"},
			wantHaveDelta:  true,
			wantDelta:      "{\n  \"Name\": \"testName\"\n}",
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
