package httpexpect

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type typedErrorNil int

func (*typedErrorNil) Error() string {
	return ""
}

func TestAssertion_Handler(t *testing.T) {
	type test struct {
		formatter *mockFormatter
		reporter  *mockReporter
		logger    *mockLogger

		handler *DefaultAssertionHandler
	}

	createTest := func(t *testing.T, enableLogger bool) test {
		var test test

		test.handler = &DefaultAssertionHandler{}

		test.formatter = newMockFormatter(t)
		test.handler.Formatter = test.formatter

		test.reporter = newMockReporter(t)
		test.handler.Reporter = test.reporter

		if enableLogger {
			test.logger = newMockLogger(t)
			test.handler.Logger = test.logger
		}

		return test
	}

	t.Run("success", func(t *testing.T) {
		test := createTest(t, true)

		test.handler.Success(&AssertionContext{
			TestName: t.Name(),
		})

		assert.Equal(t, 1, test.formatter.formattedSuccess)
		assert.Equal(t, 0, test.formatter.formattedFailure)

		assert.True(t, test.logger.logged)
		assert.False(t, test.reporter.reported)
	})

	t.Run("success_no_logger", func(t *testing.T) {
		test := createTest(t, false)

		test.handler.Success(&AssertionContext{
			TestName: t.Name(),
		})

		assert.Equal(t, 0, test.formatter.formattedSuccess)
		assert.Equal(t, 0, test.formatter.formattedFailure)

		assert.Nil(t, test.logger)
		assert.False(t, test.reporter.reported)
	})

	t.Run("failure_severity_info", func(t *testing.T) {
		test := createTest(t, true)

		test.handler.Failure(
			&AssertionContext{
				TestName: t.Name(),
			},
			&AssertionFailure{
				Type:     AssertValid,
				Severity: SeverityLog,
			})

		assert.Equal(t, 0, test.formatter.formattedSuccess)
		assert.Equal(t, 1, test.formatter.formattedFailure)

		assert.True(t, test.logger.logged)
		assert.False(t, test.reporter.reported)
	})

	t.Run("failure_severity_info_no_logger", func(t *testing.T) {
		test := createTest(t, false)

		test.handler.Failure(
			&AssertionContext{
				TestName: t.Name(),
			},
			&AssertionFailure{
				Type:     AssertValid,
				Severity: SeverityLog,
			})

		assert.Equal(t, 0, test.formatter.formattedSuccess)
		assert.Equal(t, 0, test.formatter.formattedFailure)

		assert.Nil(t, test.logger)
		assert.False(t, test.reporter.reported)
	})

	t.Run("failure_severity_error", func(t *testing.T) {
		test := createTest(t, true)

		test.handler.Failure(
			&AssertionContext{
				TestName: t.Name(),
			},
			&AssertionFailure{
				Type:     AssertValid,
				Severity: SeverityError,
			})

		assert.Equal(t, 0, test.formatter.formattedSuccess)
		assert.Equal(t, 1, test.formatter.formattedFailure)

		assert.False(t, test.logger.logged)
		assert.True(t, test.reporter.reported)
	})

	t.Run("failure_severity_error_no_logger", func(t *testing.T) {
		test := createTest(t, false)

		test.handler.Failure(
			&AssertionContext{
				TestName: t.Name(),
			},
			&AssertionFailure{
				Type:     AssertValid,
				Severity: SeverityError,
			})

		assert.Equal(t, 0, test.formatter.formattedSuccess)
		assert.Equal(t, 1, test.formatter.formattedFailure)

		assert.Nil(t, test.logger)
		assert.True(t, test.reporter.reported)
	})
}

func TestAssertion_HandlerPanics(t *testing.T) {
	t.Run("success_nil_Formatter", func(t *testing.T) {
		handler := &DefaultAssertionHandler{
			Formatter: nil,
			Reporter:  newMockReporter(t),
			Logger:    newMockLogger(t),
		}

		assert.Panics(t, func() {
			handler.Success(&AssertionContext{
				TestName: t.Name(),
			})
		})
	})

	t.Run("failure_nil_Formatter", func(t *testing.T) {
		handler := &DefaultAssertionHandler{
			Formatter: nil,
			Reporter:  newMockReporter(t),
			Logger:    newMockLogger(t),
		}

		assert.Panics(t, func() {
			handler.Failure(
				&AssertionContext{
					TestName: t.Name(),
				},
				&AssertionFailure{
					Type:     AssertValid,
					Severity: SeverityError,
				})
		})
	})

	t.Run("failure_nil_Reporter", func(t *testing.T) {
		handler := &DefaultAssertionHandler{
			Formatter: newMockFormatter(t),
			Reporter:  nil,
			Logger:    newMockLogger(t),
		}

		assert.Panics(t, func() {
			handler.Failure(
				&AssertionContext{
					TestName: t.Name(),
				},
				&AssertionFailure{
					Type:     AssertValid,
					Severity: SeverityError,
				})
		})
	})
}

func TestAssertion_ValidateTraits(t *testing.T) {
	tests := []struct {
		testName          string
		errorContainsText string
		failure           AssertionFailure
		traits            fieldTraits
	}{
		{
			testName:          "bad Type",
			errorContainsText: "AssertionType",
			failure: AssertionFailure{
				Type: AssertionType(9999),
			},
			traits: fieldTraits{
				Actual: fieldRequired,
			},
		},
		{
			testName:          "required Actual",
			errorContainsText: "Actual",
			failure: AssertionFailure{
				Actual: nil,
			},
			traits: fieldTraits{
				Actual: fieldRequired,
			},
		},
		{
			testName:          "denied Actual",
			errorContainsText: "Actual",
			failure: AssertionFailure{
				Actual: &AssertionValue{},
			},
			traits: fieldTraits{
				Actual: fieldDenied,
			},
		},
		{
			testName:          "required Expected",
			errorContainsText: "Expected",
			failure: AssertionFailure{
				Expected: nil,
			},
			traits: fieldTraits{
				Expected: fieldRequired,
			},
		},
		{
			testName:          "denied Expected",
			errorContainsText: "Expected",
			failure: AssertionFailure{
				Expected: &AssertionValue{},
			},
			traits: fieldTraits{
				Expected: fieldDenied,
			},
		},
		{
			testName:          "required AssertionRange",
			errorContainsText: "Expected",
			failure: AssertionFailure{
				Expected: nil,
			},
			traits: fieldTraits{
				Range: fieldRequired,
			},
		},
		{
			testName:          "AssertionRange should not be pointer",
			errorContainsText: "AssertionRange",
			failure: AssertionFailure{
				Expected: &AssertionValue{
					Value: &AssertionRange{},
				},
			},
			traits: fieldTraits{
				Range: fieldRequired,
			},
		},
		{
			testName:          "AssertionRange.Min should not be nil",
			errorContainsText: "AssertionRange",
			failure: AssertionFailure{
				Expected: &AssertionValue{
					Value: AssertionRange{
						Max: 1,
					},
				},
			},
			traits: fieldTraits{
				Range: fieldRequired,
			},
		},
		{
			testName:          "AssertionRange.Max should not be nil",
			errorContainsText: "AssertionRange",
			failure: AssertionFailure{
				Expected: &AssertionValue{
					Value: AssertionRange{
						Min: 1,
					},
				},
			},
			traits: fieldTraits{
				Range: fieldRequired,
			},
		},
		{
			testName:          "required AssertionList",
			errorContainsText: "Expected",
			failure: AssertionFailure{
				Expected: nil,
			},
			traits: fieldTraits{
				List: fieldRequired,
			},
		},
		{
			testName:          "AssertionList should not be pointer",
			errorContainsText: "AssertionValue",
			failure: AssertionFailure{
				Expected: &AssertionValue{
					Value: &AssertionList{},
				},
			},
			traits: fieldTraits{
				List: fieldRequired,
			},
		},
		{
			testName:          "AssertionList should be not nil",
			errorContainsText: "AssertionList",
			failure: AssertionFailure{
				Expected: &AssertionValue{
					Value: AssertionList{},
				},
			},
			traits: fieldTraits{
				List: fieldRequired,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			err := validateTraits(&test.failure, test.traits)
			require.Error(t, err)
			require.Contains(t, err.Error(), test.errorContainsText)
		})
	}

	t.Run("no error", func(t *testing.T) {
		err := validateTraits(&AssertionFailure{}, fieldTraits{})
		require.Nil(t, err)
	})

	t.Run("panic unsupported", func(t *testing.T) {
		assert.Panics(t, func() {
			err := validateTraits(&AssertionFailure{}, fieldTraits{List: fieldDenied})
			if err != nil {
				return
			}
		})

		assert.Panics(t, func() {
			err := validateTraits(&AssertionFailure{}, fieldTraits{Range: fieldDenied})
			if err != nil {
				return
			}
		})
	})

}

func TestAssertion_ValidateAssertion(t *testing.T) {
	var tnil *typedErrorNil
	var tnilPtr error = tnil

	assert.Nil(t, tnilPtr)
	assert.NotEqual(t, nil, tnilPtr)

	tests := []struct {
		testName          string
		errorContainsText string
		input             AssertionFailure
	}{
		{
			testName:          "bad Type",
			errorContainsText: "AssertionType",
			input: AssertionFailure{
				Type: AssertionType(9999),
				Errors: []error{
					errors.New("test"),
				},
			},
		},
		{
			testName:          "nil Errors",
			errorContainsText: "Errors",
			input: AssertionFailure{
				Type:   AssertOperation,
				Errors: nil,
			},
		},
		{
			testName:          "empty Errors",
			errorContainsText: "Errors",
			input: AssertionFailure{
				Type:   AssertOperation,
				Errors: []error{},
			},
		},
		{
			testName:          "nil in Errors",
			errorContainsText: "Errors",
			input: AssertionFailure{
				Type:   AssertOperation,
				Errors: []error{nil},
			},
		},
		{
			testName:          "typed nil in Errors",
			errorContainsText: "Errors",
			input: AssertionFailure{
				Type:   AssertOperation,
				Errors: []error{tnilPtr},
			},
		},
		{
			testName:          "denied Actual",
			errorContainsText: "Actual",
			input: AssertionFailure{
				Type: AssertOperation,
				Errors: []error{
					errors.New("test"),
				},
				Actual: &AssertionValue{},
			},
		},
		{
			testName:          "denied Expected",
			errorContainsText: "Expected",
			input: AssertionFailure{
				Type: AssertOperation,
				Errors: []error{
					errors.New("test"),
				},
				Expected: &AssertionValue{},
			},
		},
		{
			testName:          "required Actual and denied Expected",
			errorContainsText: "",
			input: AssertionFailure{
				Type: AssertType,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: &AssertionValue{},
			},
		},
		{
			testName:          "missing Actual and denied Expected",
			errorContainsText: "",
			input: AssertionFailure{
				Type: AssertType,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   nil,
				Expected: &AssertionValue{},
			},
		},
		{
			testName:          "missing Actual",
			errorContainsText: "Actual",
			input: AssertionFailure{
				Type: AssertEqual,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   nil,
				Expected: &AssertionValue{},
			},
		},
		{
			testName:          "missing Expected",
			errorContainsText: "Expected",
			input: AssertionFailure{
				Type: AssertEqual,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: nil,
			},
		},
		{
			testName:          "missing Actual and Expected",
			errorContainsText: "",
			input: AssertionFailure{
				Type: AssertEqual,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   nil,
				Expected: nil,
			},
		},
		{
			testName:          "Range is nil",
			errorContainsText: "AssertionRange",
			input: AssertionFailure{
				Type: AssertInRange,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: &AssertionValue{},
			},
		},
		{
			testName:          "Range has wrong type",
			errorContainsText: "AssertionRange",
			input: AssertionFailure{
				Type: AssertInRange,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: &AssertionValue{"test"},
			},
		},
		{
			testName:          "Range is pointer",
			errorContainsText: "AssertionRange",
			input: AssertionFailure{
				Type: AssertInRange,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: &AssertionValue{&AssertionRange{Min: 0, Max: 0}},
			},
		},
		{
			testName:          "Range Min is nil",
			errorContainsText: "Min",
			input: AssertionFailure{
				Type: AssertInRange,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: &AssertionValue{AssertionRange{nil, 123}},
			},
		},
		{
			testName:          "Range Max is nil",
			errorContainsText: "Max",
			input: AssertionFailure{
				Type: AssertInRange,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: &AssertionValue{AssertionRange{123, nil}},
			},
		},
		{
			testName:          "Range Min and Max are nil",
			errorContainsText: "",
			input: AssertionFailure{
				Type: AssertInRange,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: &AssertionValue{AssertionRange{nil, nil}},
			},
		},
		{
			testName:          "List is nil",
			errorContainsText: "AssertionList",
			input: AssertionFailure{
				Type: AssertBelongs,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: &AssertionValue{},
			},
		},
		{
			testName:          "List has wrong type",
			errorContainsText: "AssertionList",
			input: AssertionFailure{
				Type: AssertBelongs,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: &AssertionValue{"test"},
			},
		},
		{
			testName:          "List is pointer",
			errorContainsText: "AssertionList",
			input: AssertionFailure{
				Type: AssertBelongs,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: &AssertionValue{&AssertionList{1}},
			},
		},
		{
			testName:          "List is typed nil",
			errorContainsText: "AssertionList",
			input: AssertionFailure{
				Type: AssertBelongs,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: &AssertionValue{AssertionList(nil)},
			},
		},
		{
			testName:          "List is empty",
			errorContainsText: "AssertionList",
			input: AssertionFailure{
				Type: AssertBelongs,
				Errors: []error{
					errors.New("test"),
				},
				Actual:   &AssertionValue{},
				Expected: &AssertionValue{AssertionList{}},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.testName, func(t *testing.T) {
			err := validateAssertion(&test.input)
			require.Error(t, err)
			require.Contains(t, err.Error(), test.errorContainsText)
		})
	}
}

func TestAssertion_Strings(t *testing.T) {
	t.Run("AssertionType", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			assert.NotEmpty(t, AssertionType(i).String())
		}
	})

	t.Run("AssertionSeverity", func(t *testing.T) {
		for i := 0; i < 100; i++ {
			assert.NotEmpty(t, AssertionSeverity(i).String())
		}
	})
}
