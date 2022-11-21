package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAssertionHandler(t *testing.T) {
	var fmt *mockFormatter
	var rep *mockReporter
	var log *mockLogger

	init := func(t *testing.T) AssertionHandler {
		fmt = newMockFormatter(t)
		rep = newMockReporter(t)
		log = newMockLogger(t)

		return &DefaultAssertionHandler{
			Formatter: fmt,
			Reporter:  rep,
			Logger:    log,
		}
	}

	t.Run("Success", func(t *testing.T) {
		dah := init(t)

		dah.Success(&AssertionContext{
			TestName: t.Name(),
		})

		assert.Equal(t, 1, fmt.formattedSuccess)
		assert.Equal(t, 0, fmt.formattedFailure)

		assert.True(t, log.logged)
		assert.False(t, rep.reported)
	})

	t.Run("Failure/NonFatal", func(t *testing.T) {
		dah := init(t)

		dah.Failure(
			&AssertionContext{
				TestName: t.Name(),
			},
			&AssertionFailure{
				Type:    AssertValid,
				IsFatal: false,
			})

		assert.Equal(t, 0, fmt.formattedSuccess)
		assert.Equal(t, 1, fmt.formattedFailure)

		assert.True(t, log.logged)
		assert.False(t, rep.reported)
	})

	t.Run("Failure/Fatal", func(t *testing.T) {
		dah := init(t)

		dah.Failure(
			&AssertionContext{
				TestName: t.Name(),
			},
			&AssertionFailure{
				Type:    AssertValid,
				IsFatal: true,
			})

		assert.Equal(t, 0, fmt.formattedSuccess)
		assert.Equal(t, 1, fmt.formattedFailure)

		assert.False(t, log.logged)
		assert.True(t, rep.reported)
	})
}
