package httpexpect

// Failure contains information about failed assertion.
// It will be passed to Formatter when an assertion fails.
type Failure struct {
	// Original Error while performing an assertion
	err           error
	assertionName string
	actual        interface{}
	expected      interface{}
}
