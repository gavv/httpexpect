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
	FailureAssertUNDEFINED // used when OriginalError must be used.
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
	OriginalError    error
	CumulativeErrors []error
	AssertionName    string
	Actual           interface{}
	Expected         interface{}
	ExpectedInRange  []interface{} // [min, max]
	ExpectedDelta    interface{}
	AssertType       FailureType
}

func newErrorFailure(err error) Failure {
	return Failure{OriginalError: err, AssertType: FailureAssertUNDEFINED}
}
