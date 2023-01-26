package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockT struct {
	testing.T
}

var fatalfInvoked = false

func (m *mockT) Fatalf(format string, args ...interface{}) {
	fatalfInvoked = true
}

func TestFatalReporter(t *testing.T) {
	mockBackend := &mockT{}
	reporter := NewFatalReporter(mockBackend)

	t.Run("Test Errorf", func(t *testing.T) {
		reporter.Errorf("Test failed with backend: %v", mockBackend)
		if !mockBackend.Failed() {
			assert.True(t, fatalfInvoked)
		}
	})
}
