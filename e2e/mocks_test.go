package e2e

type mockReporter struct {
	failed bool
}

// Errorf implements Reporter.Errorf.
func (r *mockReporter) Errorf(message string, args ...interface{}) {
	r.failed = true
}
