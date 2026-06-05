package e2e

type mockReporter struct {
	failed bool
}

// Errorf implements Reporter.Errorf.
func (r *mockReporter) Errorf(_ string, _ ...interface{}) {
	r.failed = true
}
