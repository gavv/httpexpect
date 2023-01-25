package httpexpect

import (
	"testing"
)

type mockT struct {
	testing.T
}

func TestFatalReporter(t *testing.T) {
	mockBackend := &mockT{}
	reporter := NewFatalReporter(mockBackend)

	t.Run("Test Errorf", func(t *testing.T) {
		reporter.Errorf("Test failed with backend: %v", mockBackend)
		if !mockBackend.Failed() {
			t.Error("Test failed")
		}
	})
}
