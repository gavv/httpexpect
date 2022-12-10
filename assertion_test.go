package httpexpect

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAssertionHandler(t *testing.T) {
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

func TestAssertionValidation(t *testing.T) {
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
			testName:          "empty Errros",
			errorContainsText: "Errors",
			input: AssertionFailure{
				Type:   AssertOperation,
				Errors: []error{},
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

func TestAssertionStrings(t *testing.T) {
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
