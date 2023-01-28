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

	reporter.Errorf("test")
	assert.True(t, mockBackend.fatalfInvoked)
}

type mockAssertT struct {
	errorfInvoked bool
}

func (m *mockAssertT) Errorf(format string, args ...interface{}) {
	m.errorfInvoked = true
}

func TestReporter_AssertReporter(t *testing.T) {
	mockBackend := &mockAssertT{}
	reporter := NewAssertReporter(mockBackend)

	reporter.Errorf("test")
	assert.True(t, mockBackend.errorfInvoked)
}

type mockRequireT struct {
	testing.T
	failNowInvoked bool
}

func (m *mockRequireT) FailNow() {
	m.failNowInvoked = true
}

func TestReporter_RequireReporter(t *testing.T) {
	mockBackend := &mockRequireT{}
	reporter := NewRequireReporter(mockBackend)

	reporter.Errorf("test")
	assert.True(t, mockBackend.failNowInvoked)
}
