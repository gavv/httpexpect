package httpexpect

import (
	"testing"
)

func TestFatalReporter(t *testing.T) {
	mockBackend := &testing.T{}
	reporter := NewFatalReporter(mockBackend)

	t.Run("Test Errorf", func(t *testing.T) {
		reporter.Errorf("Test failed with backend: %v", mockBackend)
		if !mockBackend.Failed() {
			t.Error("Test failed")
		}
	})
}
