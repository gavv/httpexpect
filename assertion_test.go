package httpexpect

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	t.Run("success, no logger", func(t *testing.T) {
		test := createTest(t, false)

		test.handler.Success(&AssertionContext{
			TestName: t.Name(),
		})

		assert.Equal(t, 0, test.formatter.formattedSuccess)
		assert.Equal(t, 0, test.formatter.formattedFailure)

		assert.Nil(t, test.logger)
		assert.False(t, test.reporter.reported)
	})

	t.Run("failure, severity info", func(t *testing.T) {
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

	t.Run("failure, severity info, no logger", func(t *testing.T) {
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

	t.Run("failure, severity error", func(t *testing.T) {
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

	t.Run("failure, severity error, no logger", func(t *testing.T) {
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
	t.Run("success, nil Formatter", func(t *testing.T) {
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

	t.Run("failure, nil Formatter", func(t *testing.T) {
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

	t.Run("failure, nil Reporter", func(t *testing.T) {
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
	cases := []struct {
		name              string
		errorContainsText string
		failure           AssertionFailure
		traits            fieldTraits
	}{
		{
			name:              "bad Type",
			errorContainsText: "AssertionType",
			failure: AssertionFailure{
				Type: AssertionType(9999),
			},
			traits: fieldTraits{
				Actual: fieldRequired,
			},
		},
		{
			name:              "required Actual",
			errorContainsText: "Actual",
			failure: AssertionFailure{
				Actual: nil,
			},
			traits: fieldTraits{
				Actual: fieldRequired,
			},
		},
		{
			name:              "denied Actual",
			errorContainsText: "Actual",
			failure: AssertionFailure{
				Actual: &AssertionValue{},
			},
			traits: fieldTraits{
				Actual: fieldDenied,
			},
		},
		{
			name:              "required Expected",
			errorContainsText: "Expected",
			failure: AssertionFailure{
				Expected: nil,
			},
			traits: fieldTraits{
				Expected: fieldRequired,
			},
		},
		{
			name:              "denied Expected",
			errorContainsText: "Expected",
			failure: AssertionFailure{
				Expected: &AssertionValue{},
			},
			traits: fieldTraits{
				Expected: fieldDenied,
			},
		},
		{
			name:              "required AssertionRange",
			errorContainsText: "Expected",
			failure: AssertionFailure{
				Expected: nil,
			},
			traits: fieldTraits{
				Range: fieldRequired,
			},
		},
		{
			name:              "AssertionRange should not be pointer",
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
			name:              "AssertionRange.Min should not be nil",
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
			name:              "AssertionRange.Max should not be nil",
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
			name:              "required AssertionList",
			errorContainsText: "Expected",
			failure: AssertionFailure{
				Expected: nil,
			},
			traits: fieldTraits{
				List: fieldRequired,
			},
		},
		{
			name:              "AssertionList should not be pointer",
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
			name:              "AssertionList should be not nil",
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

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateTraits(&tc.failure, tc.traits)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errorContainsText)
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
	var mErr *mockError
	var mErrPtr error = mErr

	assert.Nil(t, mErrPtr)
	assert.NotEqual(t, nil, mErrPtr)

	cases := []struct {
		name              string
		errorContainsText string
		input             AssertionFailure
	}{
		{
			name:              "bad Type",
			errorContainsText: "AssertionType",
			input: AssertionFailure{
				Type: AssertionType(9999),
				Errors: []error{
					errors.New("test"),
				},
			},
		},
		{
			name:              "nil Errors",
			errorContainsText: "Errors",
			input: AssertionFailure{
				Type:   AssertOperation,
				Errors: nil,
			},
		},
		{
			name:              "empty Errors",
			errorContainsText: "Errors",
			input: AssertionFailure{
				Type:   AssertOperation,
				Errors: []error{},
			},
		},
		{
			name:              "nil in Errors",
			errorContainsText: "Errors",
			input: AssertionFailure{
				Type:   AssertOperation,
				Errors: []error{nil},
			},
		},
		{
			name:              "typed nil in Errors",
			errorContainsText: "Errors",
			input: AssertionFailure{
				Type:   AssertOperation,
				Errors: []error{mErrPtr},
			},
		},
		{
			name:              "denied Actual",
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
			name:              "denied Expected",
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
			name:              "required Actual and denied Expected",
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
			name:              "missing Actual and denied Expected",
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
			name:              "missing Actual",
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
			name:              "missing Expected",
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
			name:              "missing Actual and Expected",
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
			name:              "Range is nil",
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
			name:              "Range has wrong type",
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
			name:              "Range is pointer",
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
			name:              "Range Min is nil",
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
			name:              "Range Max is nil",
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
			name:              "Range Min and Max are nil",
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
			name:              "List is nil",
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
			name:              "List has wrong type",
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
			name:              "List is pointer",
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
			name:              "List is typed nil",
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
			name:              "List is empty",
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

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			err := validateAssertion(&tc.input)
			require.Error(t, err)
			require.Contains(t, err.Error(), tc.errorContainsText)
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
