package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAssertionHandler(t *testing.T) {
	var rep *mockReporter
	var fmt *mockFormatter

	init := func(t *testing.T) AssertionHandler {
		rep = newMockReporter(t)
		fmt = newMockFormatter(t)
		return &DefaultAssertionHandler{
			Reporter:  rep,
			Formatter: fmt,
		}
	}

	t.Run("Success", func(t *testing.T) {
		dah := init(t)

		dah.Success(&AssertionContext{
			TestName: t.Name(),
		})

		assert.False(t, rep.reported)
		assert.Zero(t, fmt.formattedFailure)
	})

	t.Run("Failure", func(t *testing.T) {
		dah := init(t)

		dah.Failure(
			&AssertionContext{
				TestName: t.Name(),
			},
			&AssertionFailure{
				Type: AssertValid,
			})

		assert.Equal(t, 1, fmt.formattedFailure)
		assert.Zero(t, fmt.formattedSuccess)
		assert.True(t, rep.reported)
	})
}
