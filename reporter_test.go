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

type mockAssertT struct {
	errorfInvoked bool
}

func (m *mockAssertT) Errorf(format string, args ...interface{}) {
	m.errorfInvoked = true
}

type mockRequireT struct {
	testing.T
	failNowInvoked bool
}

func (m *mockRequireT) FailNow() {
	m.failNowInvoked = true
}

func TestReporter_AssertReporter(t *testing.T) {
	mockBackend := &mockAssertT{}
	reporter := NewAssertReporter(mockBackend)

	reporter.Errorf("test")
	assert.True(t, mockBackend.errorfInvoked)
}

func TestReporter_RequireReporter(t *testing.T) {
	mockBackend := &mockRequireT{}
	reporter := NewRequireReporter(mockBackend)

	reporter.Errorf("test")
	assert.True(t, mockBackend.failNowInvoked)
}

func TestReporter_FatalReporter(t *testing.T) {
	mockBackend := &mockT{}
	reporter := NewFatalReporter(mockBackend)

	reporter.Errorf("test")
	assert.True(t, mockBackend.fatalfInvoked)
}

func TestReporter_PanicReporter(t *testing.T) {
	reporter := NewPanicReporter()

	assert.Panics(t, func() {
		reporter.Errorf("test")
	})
}
