package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type mockT struct {
	testing.T
	fatalfInvoked bool
}

func (m *mockT) Fatalf(format string, args ...interface{}) {
	m.fatalfInvoked = true
}

func TestReporter_FatalReporter(t *testing.T) {
	mockBackend := &mockT{}
	reporter := NewFatalReporter(mockBackend)

	t.Run("Errorf", func(t *testing.T) {
		reporter.Errorf("Test failed with backend: %v", mockBackend)
		assert.True(t, mockBackend.fatalfInvoked)
	})
}
