package httpexpect

// AssertionType defines type of performed assertion.
type AssertionType uint64

//go:generate stringer -type=AssertionType
const (
	// Check if the invocation is correct
	AssertUsage AssertionType = iota

	// Check if the operation succeeded
	AssertOperation

	// Check expression: [Actual] has valid layout
	AssertValid
	AssertNotValid

	// Check expression: [Actual] is nil
	AssertNil
	AssertNotNil

	// Check expression: [Actual] is empty
	AssertEmpty
	AssertNotEmpty

	// Check expression: [Actual] is equal to [Expected]
	// If non-zero, [Delta] specifies allowed difference between values
	AssertEqual
	AssertNotEqual

	// Check expression: [Actual] < [Expected]
	AssertLt
	// Check expression: [Actual] <= [Expected]
	AssertLe
	// Check expression: [Actual] > [Expected]
	AssertGt
	// Check expression: [Actual] >= [Expected]
	AssertGe

	// Check expression: [Actual] belongs to inclusive range [Expected]
	// [Expected] value stores AssertionRange
	AssertInRange
	AssertNotInRange

	// Check expression: [Actual] matches json schema [Expected]
	AssertMatchSchema
	AssertNotMatchSchema

	// Check expression: [Actual] matches json path [Expected]
	AssertMatchPath
	AssertNotMatchPath

	// Check expression: [Actual] matches regex [Expected]
	AssertMatchRegexp
	AssertNotMatchRegexp

	// Check expression: [Actual] contains key [Expected]
	AssertContainsKey
	AssertNotContainsKey

	// Check expression: [Actual] contains element [Expected]
	AssertContainsElement
	AssertNotContainsElement

	// Check expression: [Actual] contains subset [Expected]
	AssertContainsSubset
	AssertNotContainsSubset

	// Check expression: [Actual] belongs to list [Expected]
	// [Expected] value stores AssertionList
	AssertBelongs
	AssertNotBelongs
)

// AssertionContext provides context where the assetion happened.
type AssertionContext struct {
	// Name of the running test
	// Usually comes from testing.T
	TestName string

	// Chain of nested assertion names
	// Example value:
	//   {`Request("GET")`, `Expect()`, `JSON()`, `NotNull()`}
	Path []string

	// Request being sent
	// May be nil if request was not yet sent
	Request *Request

	// Response being matched
	// May be nil if request was not yet received
	Response *Response
}

// AssertionFailure provides detailed information about failed assertion.
//
// [Type] and [Errors] fields are set for all assertions.
// [Actual] and [Expected] fields are set only for certain assertion types.
//
// The value itself is stored in [Actual.Value] and [Expected.Value], which
// allows to distinguish whether [Actual] or [Expected] is not present at
// all, or is present but contains nil value.
//
// [Actual] stores the value being examined.
//
// Exact meaning of [Expected] depends on assertion type. It may be the value
// to which [Actual] was compared, or range to which [Actual] should belong,
// or pattern with which [Actual] should match, or element which [Actual]
// should contain, and so on.
//
// For details, see comments for corresponding AssertionType constant.
type AssertionFailure struct {
	// Type of failed assertion
	Type AssertionType

	// List of error messages
	Errors []error

	// Actually observed value
	Actual *AssertionValue

	// Expected value
	Expected *AssertionValue

	// Allowed delta between actual and expected
	Delta float64
}

// AssertionValue holds expected or actual value
type AssertionValue struct {
	Value interface{}
}

// AssertionRange holds [min; max] range for value
type AssertionRange [2]interface{}

// AssertionList holds list of allowed values
type AssertionList []interface{}

// AssertionHandler takes care of formatting and reporting test Failure or Success.
//
// You can log every performed assertion, or report only failures. You can implement
// custom formatting, for example, provide a JSON output for ulterior processing.
type AssertionHandler interface {
	Success(*AssertionContext)
	Failure(*AssertionContext, *AssertionFailure)
}

// DefaultAssertionHandler is default implementation for AssertionHandler.
//
// Uses Formatter to format failure message, Reporter to report failure messages,
// and Logger to report success messages.
//
// Logger is optional. Set it if you want to log every successful assertion.
// Reporter and Formatter are required.
type DefaultAssertionHandler struct {
	Reporter  Reporter
	Logger    Logger
	Formatter Formatter
}

// Success implements AssertionHandler.Success.
func (h *DefaultAssertionHandler) Success(ctx *AssertionContext) {
	if h.Logger == nil {
		return
	}

	msg := h.Formatter.FormatSuccess(ctx)

	h.Logger.Logf("%s", msg)
}

// Failure implements AssertionHandler.Failure.
func (h *DefaultAssertionHandler) Failure(
	ctx *AssertionContext, failure *AssertionFailure,
) {
	if h.Formatter == nil {
		panic("DefaultAssertionHandler.Formatter is nil")
	}
	if h.Reporter == nil {
		panic("DefaultAssertionHandler.Reporter is nil")
	}

	msg := h.Formatter.FormatFailure(ctx, failure)

	h.Reporter.Errorf("%s", msg)
}
