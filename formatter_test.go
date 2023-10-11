package httpexpect

import (
	"encoding/json"
	"fmt"
	"math/big"
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
	cases := []struct {
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
			wantActual:     "int(1_000_000)",
		},
		{
			name:           "AssertType float32",
			assertionType:  AssertType,
			assertionValue: float32(1_000_000),
			wantHaveActual: true,
			wantActual:     "float32(1_000_000)",
		},
		{
			name:           "AssertType float64",
			assertionType:  AssertType,
			assertionValue: float64(1_000_000),
			wantHaveActual: true,
			wantActual:     "float64(1_000_000)",
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
			wantActual:     "1_000_000",
		},
		{
			name:           "AssertValid float32",
			assertionType:  AssertValid,
			assertionValue: float32(1_000_000),
			wantHaveActual: true,
			wantActual:     "1_000_000",
		},
		{
			name:           "AssertValid float64",
			assertionType:  AssertValid,
			assertionValue: float64(1_000_000),
			wantHaveActual: true,
			wantActual:     "1_000_000",
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

	for _, tc := range cases {
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
	cases := []struct {
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
			wantExpected:     []string{"[1_000_000; 2_000_000]"},
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
			wantExpected:     []string{"[1_000_000; 2_000_000]"},
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
			wantExpected:     []string{"[1_000_000; 2_000_000]"},
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
			wantExpected:     []string{"1_000_000"},
		},
		{
			name:             "AssertMatchPath float32",
			assertionType:    AssertMatchPath,
			assertionValue:   float32(1_000_000),
			wantHaveExpected: true,
			wantExpectedKind: kindPath,
			wantExpected:     []string{"1_000_000"},
		},
		{
			name:             "AssertMatchPath float64",
			assertionType:    AssertMatchPath,
			assertionValue:   float64(1_000_000),
			wantHaveExpected: true,
			wantExpectedKind: kindPath,
			wantExpected:     []string{"1_000_000"},
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
			wantExpected:     []string{"1_000_000", "2_000_000"},
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
			wantExpected:     []string{"1_000_000", "2_000_000"},
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
			wantExpected:     []string{"1_000_000", "2_000_000"},
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

	for _, tc := range cases {
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
	cases := []struct {
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
			wantReference:     "1_000_000",
		},
		{
			name:              "float32",
			assertionValue:    float32(1_000_000),
			wantHaveReference: true,
			wantReference:     "1_000_000",
		},
		{
			name:              "float64",
			assertionValue:    float64(1_000_000),
			wantHaveReference: true,
			wantReference:     "1_000_000",
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

	for _, tc := range cases {
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
	cases := []struct {
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
			wantDelta:      "1_000_000",
		},
		{
			name:           "float32",
			assertionValue: float32(1_000_000),
			wantHaveDelta:  true,
			wantDelta:      "1_000_000",
		},
		{
			name:           "float64",
			assertionValue: float64(1_000_000),
			wantHaveDelta:  true,
			wantDelta:      "1_000_000",
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

	for _, tc := range cases {
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

func TestFormatter_FailureErrors(t *testing.T) {
	var mErr *mockError
	var mErrPtr error = mErr

	assert.Nil(t, mErrPtr)
	assert.NotEqual(t, nil, mErrPtr)

	cases := []struct {
		name     string
		errors   []error
		expected []string
	}{
		{
			name:     "nil errors slice",
			errors:   nil,
			expected: []string{},
		},
		{
			name:     "empty errors slice",
			errors:   []error{},
			expected: []string{},
		},
		{
			name:     "errors slice with nil error",
			errors:   []error{nil},
			expected: []string{},
		},
		{
			name:     "errors slice with typed nil error",
			errors:   []error{mErrPtr},
			expected: []string{},
		},
		{
			name:     "errors slice with one error",
			errors:   []error{fmt.Errorf("error message")},
			expected: []string{"error message"},
		},
		{
			name: "errors slice with multiple errors",
			errors: []error{
				fmt.Errorf("error message 1"),
				fmt.Errorf("error message 2"),
			},
			expected: []string{
				"error message 1",
				"error message 2",
			},
		},
	}

	df := &DefaultFormatter{}
	ctx := &AssertionContext{}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fl := &AssertionFailure{
				Errors: tc.errors,
			}
			fd := df.buildFormatData(ctx, fl)
			assert.Equal(t, tc.expected, fd.Errors)
		})
	}
}

func TestFormatter_FailureContext(t *testing.T) {
	ctx := &AssertionContext{
		TestName:    "MyTestName",
		RequestName: "MyRequestName",
		Path:        []string{"MyPath"},
		AliasedPath: []string{"MyAliasedPath"},
	}

	cases := []struct {
		name  string
		fmt   DefaultFormatter
		check func(t *testing.T, data *FormatData)
	}{
		{
			name: "default options",
			fmt:  DefaultFormatter{},
			check: func(t *testing.T, fd *FormatData) {
				assert.Equal(t, "MyTestName", fd.TestName)
				assert.Equal(t, "MyRequestName", fd.RequestName)
				assert.Equal(t, []string{"MyAliasedPath"}, fd.AssertPath)
			},
		},
		{
			name: "DisableNames",
			fmt: DefaultFormatter{
				DisableNames: true,
			},
			check: func(t *testing.T, fd *FormatData) {
				assert.Equal(t, "", fd.TestName)
				assert.Equal(t, "", fd.RequestName)
				assert.Equal(t, []string{"MyAliasedPath"}, fd.AssertPath)
			},
		},
		{
			name: "DisablePaths",
			fmt: DefaultFormatter{
				DisablePaths: true,
			},
			check: func(t *testing.T, fd *FormatData) {
				assert.Equal(t, "MyTestName", fd.TestName)
				assert.Equal(t, "MyRequestName", fd.RequestName)
				assert.Equal(t, []string(nil), fd.AssertPath)
			},
		},
		{
			name: "DisableAliases",
			fmt: DefaultFormatter{
				DisableAliases: true,
			},
			check: func(t *testing.T, fd *FormatData) {
				assert.Equal(t, "MyTestName", fd.TestName)
				assert.Equal(t, "MyRequestName", fd.RequestName)
				assert.Equal(t, []string{"MyPath"}, fd.AssertPath)
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			fl := &AssertionFailure{
				Type: AssertEqual,
			}
			fd := tc.fmt.buildFormatData(ctx, fl)
			tc.check(t, fd)
		})
	}
}

func TestFormatter_FloatFormat(t *testing.T) {
	cases := []struct {
		name     string
		format   FloatFormat
		value    interface{}
		wantText string
	}{
		// float32
		{
			name:     "float32 auto small exponent",
			format:   FloatFormatAuto,
			value:    float32(1.2345678),
			wantText: "1.234_567_8",
		},
		{
			name:     "float32 auto large exponent",
			format:   FloatFormatAuto,
			value:    float32(1234567.8),
			wantText: "1.234_567_8e+06",
		},
		{
			name:     "float32 decimal",
			format:   FloatFormatDecimal,
			value:    float32(1234567.8),
			wantText: "1_234_567.8",
		},
		{
			name:     "float32 scientific",
			format:   FloatFormatScientific,
			value:    float32(1.2345678),
			wantText: "1.234_567_8e+00",
		},
		// float64
		{
			name:     "float64 auto small exponent",
			format:   FloatFormatAuto,
			value:    float64(1.23456789),
			wantText: "1.234_567_89",
		},
		{
			name:     "float64 auto large exponent",
			format:   FloatFormatAuto,
			value:    float64(12345678.9),
			wantText: "1.234_567_89e+07",
		},
		{
			name:     "float64 decimal",
			format:   FloatFormatDecimal,
			value:    float64(12345678.9),
			wantText: "12_345_678.9",
		},
		{
			name:     "float64 scientific",
			format:   FloatFormatScientific,
			value:    float64(1.23456789),
			wantText: "1.234_567_89e+00",
		},
		// no fractional part
		{
			name:     "nofrac auto",
			format:   FloatFormatAuto,
			value:    float32(12345678),
			wantText: "12_345_678",
		},
		{
			name:     "nofrac decimal",
			format:   FloatFormatDecimal,
			value:    float32(12345678),
			wantText: "12_345_678",
		},
		{
			name:     "nofrac scientific",
			format:   FloatFormatScientific,
			value:    float32(12345678),
			wantText: "1.234_567_8e+07",
		},
		// integer
		{
			name:     "integer auto",
			format:   FloatFormatAuto,
			value:    int(12345678),
			wantText: "12_345_678",
		},
		{
			name:     "integer decimal",
			format:   FloatFormatDecimal,
			value:    int(12345678),
			wantText: "12_345_678",
		},
		{
			name:     "integer scientific",
			format:   FloatFormatScientific,
			value:    int(12345678),
			wantText: "12_345_678",
		},
		// slice of floats
		{
			name:     "slice of float auto",
			format:   FloatFormatAuto,
			value:    []float32{1.234, 0.0056, 78000},
			wantText: "[\n  1.234,\n  0.0056,\n  78000\n]",
		},
		{
			name:     "slice of float decimal",
			format:   FloatFormatDecimal,
			value:    []float32{1.234, 0.0056, 78000},
			wantText: "[\n  1.234,\n  0.0056,\n  78000\n]",
		},
		{
			name:     "slice of float scientific",
			format:   FloatFormatScientific,
			value:    []float32{1.234, 0.0056, 78000},
			wantText: "[\n  1.234e+00,\n  5.6e-03,\n  7.8e+04\n]",
		},
		// slice of json.Number
		{
			name:     "slice of json.Number auto",
			format:   FloatFormatAuto,
			value:    []json.Number{"12.34", ".0056", "789"},
			wantText: "[\n  12.34,\n  0.0056,\n  789\n]",
		},
		{
			name:     "slice of json.Number decimal",
			format:   FloatFormatDecimal,
			value:    []json.Number{"12.34", ".0056", "789"},
			wantText: "[\n  12.34,\n  0.0056,\n  789\n]",
		},
		{
			name:     "slice of json.Number scientific",
			format:   FloatFormatScientific,
			value:    []json.Number{"12.34", ".0056", "789"},
			wantText: "[\n  1.234e+01,\n  5.6e-03,\n  7.89e+02\n]",
		},
		// slice of big.Float
		{
			name:     "slice of big.Float auto",
			format:   FloatFormatAuto,
			value:    []*big.Float{big.NewFloat(1234.5678), big.NewFloat(0.000234)},
			wantText: "[\n  1234.5678,\n  0.000234\n]",
		},
		{
			name:     "slice of big.Float decimal",
			format:   FloatFormatDecimal,
			value:    []*big.Float{big.NewFloat(1234.5678), big.NewFloat(0.000234)},
			wantText: "[\n  1234.5678,\n  0.000234\n]",
		},
		{
			name:     "slice of big.Float scientific",
			format:   FloatFormatScientific,
			value:    []*big.Float{big.NewFloat(1234.5678), big.NewFloat(0.000234)},
			wantText: "[\n  1.2345678e+03,\n  2.34e-04\n]",
		},
		// slice of slice
		{
			name:     "slice of slice auto",
			format:   FloatFormatAuto,
			value:    []interface{}{[]float64{.01, 20}, big.NewFloat(.0025), []json.Number{"23.45"}, 10},
			wantText: "[\n  [\n    0.01,\n    20\n  ],\n  0.0025,\n  [\n    23.45\n  ],\n  10\n]",
		},
		{
			name:     "slice of slice decimal",
			format:   FloatFormatDecimal,
			value:    []interface{}{[]float64{.01, 20}, big.NewFloat(.0025), []json.Number{"23.45"}, 10},
			wantText: "[\n  [\n    0.01,\n    20\n  ],\n  0.0025,\n  [\n    23.45\n  ],\n  10\n]",
		},
		{
			name:     "slice of slice scientific",
			format:   FloatFormatScientific,
			value:    []interface{}{[]float64{.01, 20}, big.NewFloat(.0025), []json.Number{"23.45"}, 10},
			wantText: "[\n  [\n    1e-02,\n    2e+01\n  ],\n  2.5e-03,\n  [\n    2.345e+01\n  ],\n  1e+01\n]",
		},
		// map of floats
		{
			name:     "map of float auto",
			format:   FloatFormatAuto,
			value:    map[string]float32{"a": 123.45, "b": 0.00678},
			wantText: "{\n  \"a\": 123.45,\n  \"b\": 0.00678\n}",
		},
		{
			name:     "map of float decimal",
			format:   FloatFormatDecimal,
			value:    map[string]float32{"a": 123.45, "b": 0.00678},
			wantText: "{\n  \"a\": 123.45,\n  \"b\": 0.00678\n}",
		},
		{
			name:     "map of float scientific",
			format:   FloatFormatScientific,
			value:    map[string]float32{"a": 123.45, "b": 0.00678},
			wantText: "{\n  \"a\": 1.2345e+02,\n  \"b\": 6.78e-03\n}",
		},
		// map of json.Number
		{
			name:     "map of json.Number auto",
			format:   FloatFormatAuto,
			value:    map[int]json.Number{1: "0.123", 2: "45.67", 3: "100"},
			wantText: "{\n  \"1\": 0.123,\n  \"2\": 45.67,\n  \"3\": 100\n}",
		},
		{
			name:     "map of json.Number decimal",
			format:   FloatFormatDecimal,
			value:    map[int]json.Number{1: "0.123", 2: "45.67", 3: "100"},
			wantText: "{\n  \"1\": 0.123,\n  \"2\": 45.67,\n  \"3\": 100\n}",
		},
		{
			name:     "map of json.Number scientific",
			format:   FloatFormatScientific,
			value:    map[int]json.Number{1: "0.123", 2: "45.67", 3: "100"},
			wantText: "{\n  \"1\": 1.23e-01,\n  \"2\": 4.567e+01,\n  \"3\": 1e+02\n}",
		},
		// map of any
		{
			name:     "map of any auto",
			format:   FloatFormatAuto,
			value:    map[string]interface{}{"a": []float32{12.34}, "b": big.NewFloat(45.67), "c": []int{100}},
			wantText: "{\n  \"a\": [\n    12.34\n  ],\n  \"b\": 45.67,\n  \"c\": [\n    100\n  ]\n}",
		},
		{
			name:     "map of any decimal",
			format:   FloatFormatDecimal,
			value:    map[string]interface{}{"a": []float32{12.34}, "b": big.NewFloat(45.67), "c": []int{100}},
			wantText: "{\n  \"a\": [\n    12.34\n  ],\n  \"b\": 45.67,\n  \"c\": [\n    100\n  ]\n}",
		},
		{
			name:     "map of any scientific",
			format:   FloatFormatScientific,
			value:    map[string]interface{}{"a": []float32{12.34}, "b": big.NewFloat(45.67), "c": []int{100}},
			wantText: "{\n  \"a\": [\n    1.234e+01\n  ],\n  \"b\": 4.567e+01,\n  \"c\": [\n    1e+02\n  ]\n}",
		},
		// dump go values
		{
			name:     "go values auto",
			format:   FloatFormatAuto,
			value:    map[[1]int][]float64{{1}: {12345.678}},
			wantText: "map[[1]int]interface {}{\n  [1]int{\n    1,\n  }: []interface {}{\n    *httpexpect.formatNumber12_345.678,\n  },\n}",
		},
		{
			name:     "go values decimal",
			format:   FloatFormatDecimal,
			value:    map[[1]int][]float64{{1}: {12345.678}},
			wantText: "map[[1]int]interface {}{\n  [1]int{\n    1,\n  }: []interface {}{\n    *httpexpect.formatNumber12_345.678,\n  },\n}",
		},
		{
			name:     "go values scientific",
			format:   FloatFormatScientific,
			value:    map[[1]int][]float64{{1}: {12345.678}},
			wantText: "map[[1]int]interface {}{\n  [1]int{\n    1,\n  }: []interface {}{\n    *httpexpect.formatNumber1.234_567_8e+04,\n  },\n}",
		},
	}

	for _, tc := range cases {
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
	cases := []struct {
		name     string
		format   FloatFormat
		value    float64
		wantText string
	}{
		{
			name:     "auto",
			format:   FloatFormatAuto,
			value:    1.2345678,
			wantText: "1.234_567_8",
		},
		{
			name:     "decimal",
			format:   FloatFormatDecimal,
			value:    1.2345678,
			wantText: "1.234_567_8",
		},
		{
			name:     "scientific",
			format:   FloatFormatScientific,
			value:    1.2345678,
			wantText: "1.234_567_8e+00",
		},
	}

	for _, tc := range cases {
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

func TestFormatter_DigitSeparator(t *testing.T) {
	cases := []struct {
		name      string
		separator DigitSeparator
		format    FloatFormat
		value     interface{}
		wantText  string
	}{
		// types
		{
			name:      "float32",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatAuto,
			value:     float32(1.23456),
			wantText:  "1.234_56",
		},
		{
			name:      "float64",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatAuto,
			value:     float64(1.23456789),
			wantText:  "1.234_567_89",
		},
		{
			name:      "int32",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatAuto,
			value:     int32(12345678),
			wantText:  "12_345_678",
		},
		{
			name:      "int64",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatAuto,
			value:     int64(12345678),
			wantText:  "12_345_678",
		},
		// components
		{
			name:      "int part, decimal",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatDecimal,
			value:     float64(12345678),
			wantText:  "12_345_678",
		},
		{
			name:      "int part, scientific",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatScientific,
			value:     float64(12345678),
			wantText:  "1.234_567_8e+07",
		},
		{
			name:      "sign part, int part, decimal",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatDecimal,
			value:     float64(-12345678),
			wantText:  "-12_345_678",
		},
		{
			name:      "sign part, int part, scientific",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatScientific,
			value:     float64(-12345678),
			wantText:  "-1.234_567_8e+07",
		},
		{
			name:      "int part, frac part, decimal",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatDecimal,
			value:     float64(12345678.12345678),
			wantText:  "12_345_678.123_456_78",
		},
		{
			name:      "int part, frac part, scientific",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatScientific,
			value:     float64(12345678.12345678),
			wantText:  "1.234_567_812_345_678e+07",
		},
		{
			name:      "sign part, int part, frac part, decimal",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatDecimal,
			value:     float64(-12345678.12345678),
			wantText:  "-12_345_678.123_456_78",
		},
		{
			name:      "sign part, int part, frac part, scientific",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatScientific,
			value:     float64(-12345678.12345678),
			wantText:  "-1.234_567_812_345_678e+07",
		},
		// edge cases
		{
			name:      "int part, 3 digits",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatDecimal,
			value:     float64(123),
			wantText:  "123",
		},
		{
			name:      "int part, multiple of 3",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatDecimal,
			value:     float64(123456),
			wantText:  "123_456",
		},
		{
			name:      "frac part, 3 digits",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatDecimal,
			value:     float64(0.123),
			wantText:  "0.123",
		},
		{
			name:      "frac part, multiple of 3",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatDecimal,
			value:     float64(0.123456),
			wantText:  "0.123_456",
		},
		{
			name:      "int and frac part, multiple of 3",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatDecimal,
			value:     float64(123456.123456),
			wantText:  "123_456.123_456",
		},
		// separators
		{
			name:      "underscore",
			separator: DigitSeparatorUnderscore,
			format:    FloatFormatDecimal,
			value:     float64(12345678),
			wantText:  "12_345_678",
		},
		{
			name:      "comma",
			separator: DigitSeparatorComma,
			format:    FloatFormatDecimal,
			value:     float64(12345678),
			wantText:  "12,345,678",
		},
		{
			name:      "apostrophe",
			separator: DigitSeparatorApostrophe,
			format:    FloatFormatDecimal,
			value:     float64(12345678),
			wantText:  "12'345'678",
		},
		{
			name:      "none",
			separator: DigitSeparatorNone,
			format:    FloatFormatDecimal,
			value:     float64(12345678),
			wantText:  "12345678",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			formatter := DefaultFormatter{
				DigitSeparator: tc.separator,
				FloatFormat:    tc.format,
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

func TestFormatter_StacktraceMode(t *testing.T) {
	cases := []struct {
		name string
		mode StacktraceMode
		want bool
	}{
		{
			name: "disabled",
			mode: StacktraceModeDisabled,
			want: false,
		},
		{
			name: "standard",
			mode: StacktraceModeStandard,
			want: true,
		},
		{
			name: "compact",
			mode: StacktraceModeCompact,
			want: true,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			f := &DefaultFormatter{
				StacktraceMode: tc.mode,
			}
			fd := f.buildFormatData(&AssertionContext{}, &AssertionFailure{
				Stacktrace: stacktrace(),
			})

			if tc.want {
				require.GreaterOrEqual(t, len(fd.Stacktrace), 1)
				assert.Contains(t, fd.Stacktrace[0], "TestFormatter_StacktraceMode.func")
				assert.Contains(t, fd.Stacktrace[0], "formatter_test.go")
				assert.Contains(t, fd.Stacktrace[0], "github.com/gavv/httpexpect")
			} else {
				assert.NotNil(t, fd.Stacktrace)
				assert.Equal(t, 0, len(fd.Stacktrace))
			}
		})
	}
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
