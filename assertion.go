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
	// [Delta] specifies allowed difference between values (may be zero)
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
	// [Expected] stores AssertionRange with Min and Max values
	AssertInRange
	AssertNotInRange

	// Check expression: [Actual] matches json schema [Expected]
	// [Expected] stores map with parsed schema or string with schema uri
	AssertMatchSchema
	AssertNotMatchSchema

	// Check expression: [Actual] matches json path [Expected]
	// [Expected] stores a string with json path
	AssertMatchPath
	AssertNotMatchPath

	// Check expression: [Actual] matches regex [Expected]
	// [Expected] stores a string with regular expression
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
	// [Expected] stores AssertionList with allowed values
	AssertBelongs
	AssertNotBelongs
)

// AssertionContext provides context where the assetion happened.
type AssertionContext struct {
	// Name of the running test
	// Usually comes from testing.T
	TestName string

	// Name of request being sent
	// Comes from Request.WithName()
	RequestName string

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

	// Defines if failure should be reported as fatal
	IsFatal bool

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

// AssertionRange holds inclusive range for allowed values
type AssertionRange struct {
	Min interface{}
	Max interface{}
}

// AssertionList holds list of allowed values
type AssertionList []interface{}

// AssertionHandler takes care of formatting and reporting test Failure or Success.
//
// You can log every performed assertion, or report only failures. You can implement
// custom formatting, for example, provide a JSON output for ulterior processing.
type AssertionHandler interface {
	// Invoked every time when an assertion succeeded.
	// May print failure to log or ignore it.
	Success(*AssertionContext)

	// Invoked every time when an assertion failed.
	// If Failure.IsFatal is false, may print failure to log or ignore it.
	// If Failure.IsFatal is true, should report failure to testing suite.
	Failure(*AssertionContext, *AssertionFailure)
}

// DefaultAssertionHandler is default implementation for AssertionHandler.
//
// - Formatter is used to format success and failure messages
// - Reporter is used to report formatted fatal failure messages
// - Logger is used to print formatted success and non-fatal failure messages
//
// Formatter and Reporter are required. Logger is optional.
// By default httpexpect creates DefaultAssertionHandler without Logger.
type DefaultAssertionHandler struct {
	Formatter Formatter
	Reporter  Reporter
	Logger    Logger
}

// Success implements AssertionHandler.Success.
func (h *DefaultAssertionHandler) Success(ctx *AssertionContext) {
	if h.Formatter == nil {
		panic("DefaultAssertionHandler.Formatter is nil")
	}

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

	if failure.IsFatal {
		if h.Reporter == nil {
			panic("DefaultAssertionHandler.Reporter is nil")
		}

		msg := h.Formatter.FormatFailure(ctx, failure)

		h.Reporter.Errorf("%s", msg)
	} else {
		if h.Logger == nil {
			return
		}

		msg := h.Formatter.FormatFailure(ctx, failure)

		h.Logger.Logf("%s", msg)
	}
}
