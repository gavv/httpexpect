package httpexpect

type failureAssertType uint64

const (
	// FIXME: use stringer to provide a string representation and allow those const to be used publicly?
	failureAssertEqual failureAssertType = iota
	failureAssertNotEqual
	failureAssertLt
	failureAssertLe
	failureAssertGe
	failureAssertGt
	failureAssertInRange
	failureAssertEmpty
	failureAssertNotEmpty
	failureAssertFirst
	failureAssertLast
	failureAssertContainsOnly
	failureAssertContains
	failureAssertNotContains
	failureAssertNotNil
	failureAssertNil
	failureAssertOutOfBounds
	failureAssertIsSet
	failureAssertNotSet
	failureAssertUNDEFINED // used when err must be used.
	failureAssertJsonSchema
	failureAssertMatchName
	failureAssertMatchEmpty
	failureAssertMatchNotEmpty
	failureAssertMatchOutOfBounds
	failureAssertMatchValues
	failureAssertMatchNotValues
	failureAssertEqualDelta
	failureAssertNotEqualDelta
	failureAssertKey
	failureAssertContainsKey
	failureAssertNotContainsKey
	failureAssertContainsMap
	failureAssertNotContainsMap
	failureAssertHTTPStatusRange
	failureAssertCookie
	failureAssertMatchRe
	failureAssertNotMatchRe
	failureAssertBadValueType
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
	assertType       failureAssertType
}

func NewErrorFailure(err error) Failure {
	return Failure{err: err, assertType: failureAssertUNDEFINED}
}
