package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AssertReporter implements Reporter interface using `testify/assert'
// package. Failures are non-fatal with this reporter.
type AssertReporter struct {
	backend *assert.Assertions
}

// NewAssertReporter returns a new AssertReporter object.
func NewAssertReporter(t assert.TestingT) *AssertReporter {
	return &AssertReporter{assert.New(t)}
}

// Errorf implements Reporter.Errorf.
func (r *AssertReporter) Errorf(message string, args ...interface{}) {
	r.backend.Fail(message, args...)
}

// RequireReporter implements Reporter interface using `testify/require'
// package. Failures are fatal with this reporter.
type RequireReporter struct {
	backend *require.Assertions
}

// NewRequireReporter returns a new RequireReporter object.
func NewRequireReporter(t require.TestingT) *RequireReporter {
	return &RequireReporter{require.New(t)}
}

// Errorf implements Reporter.Errorf.
func (r *RequireReporter) Errorf(message string, args ...interface{}) {
	r.backend.FailNow(message, args...)
}

// FatalReporter is a struct that implements the Reporter interface
// and calls t.Fatalf() when a test fails.
type FatalReporter struct {
	backend testing.TB
}

// PanicReporter is similar to FatalReporter, but instead of calling t.Fatalf it just calls panic.
type PanicReporter struct {
	backend testing.TB
}

// NewPanicReporter returns a new PanicReporter object.
func NewPanicReporter(t testing.TB) *PanicReporter {
	return &PanicReporter{t}
}

// Errorf implements Reporter.Errorf.
// it should only call panic
func (r *PanicReporter) Errorf(message string, args ...interface{}) {
	r.backend.Fatalf(message, args...)
}

// NewFatalReporter returns a new FatalReporter object.
func NewFatalReporter(t testing.TB) *FatalReporter {
	return &FatalReporter{t}
}

// Errorf implements Reporter.Errorf.
func (r *FatalReporter) Errorf(message string, args ...interface{}) {
	r.backend.Fatalf(message, args...)
}
