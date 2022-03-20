package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultAssertionHandler(t *testing.T) {
	var mr *mockReporter
	var mf *mockFormatter

	init := func(t *testing.T) AssertionHandler {
		mr = newMockReporter(t)
		mf = newMockFormatter(t)
		return DefaultAssertionHandler{
			Reporter:  mr,
			Formatter: mf,
		}
	}

	t.Run("Errorf", func(t *testing.T) {
		dah := init(t)
		dah.Errorf("msg", "arg1")
		assert.Zero(t, mf.formattedFailure)
		assert.Zero(t, mf.formattedSuccess)
		assert.True(t, mr.reported)
	})

	t.Run("Failure", func(t *testing.T) {
		dah := init(t)
		dah.Failure(&Context{TestName: t.Name()}, Failure{AssertionName: "Failure"})
		assert.Equal(t, 1, mf.formattedFailure)
		assert.Zero(t, mf.formattedSuccess)
		assert.True(t, mr.reported)
	})

	t.Run("Success", func(t *testing.T) {
		dah := init(t)
		dah.Success(&Context{TestName: t.Name()})
		assert.False(t, mr.reported)
		assert.Zero(t, mf.formattedFailure)
		// TODO: verify when used
		// assert.Equal(t, 1, mf.formattedSuccess)
	})
}
