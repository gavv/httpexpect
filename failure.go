package httpexpect

type FailureType uint64

const (
	FailureAssertEqual FailureType = iota
	FailureAssertNotEqual
	FailureAssertLt
	FailureAssertLe
	FailureAssertGe
	FailureAssertGt
	FailureAssertInRange
	FailureAssertEmpty
	FailureAssertNotEmpty
	FailureAssertContainsOnly
	FailureAssertContains
	FailureAssertNotContains
	FailureAssertNotNil
	FailureAssertNil
	FailureAssertOutOfBounds
	FailureAssertUNDEFINED // used when err must be used.
	FailureAssertJsonSchema
	FailureAssertEqualDelta
	FailureAssertNotEqualDelta
	FailureAssertKey
	FailureAssertHTTPStatusRange
	FailureAssertCookie
	FailureAssertMatchRe
	FailureAssertNotMatchRe
	FailureAssertBadType
	FailureInvalidInput
)

// Failure contains information about failed assertion.
// It will be passed to Formatter when an assertion fails.
type Failure struct {
	// Original Error while performing an assertion
	err              error
	cumulativeErrors []error
	assertionName    string
	actual           interface{}
	expected         interface{}
	expectedInRange  []interface{} // [min, max]
	expectedDelta    interface{}
	assertType       FailureType
}

func newErrorFailure(err error) Failure {
	return Failure{err: err, assertType: FailureAssertUNDEFINED}
}
