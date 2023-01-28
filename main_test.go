package httpexpect

import (
	"os"
	"testing"
)

func TestMain(m *testing.M) {
	// all chains should panic on misuse
	chainValidation = true

	os.Exit(m.Run())
}
