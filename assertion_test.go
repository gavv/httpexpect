package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
