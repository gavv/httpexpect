package httpexpect

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// Configure chain to panic if it receives an illformed assertion, to detect
	// this kind of bugs when running tests.
	// This option is not enabled outside of our own tests. Illformed assertions
	// indicates non-critical bugs, and in most cases it's still possible to format
	// and show assertion to user, which is better than refusing to work.
	chainAssertionValidation = true

	os.Exit(m.Run())
}
