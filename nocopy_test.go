package httpexpect

import (
	"testing"
)

func TestNoCopy(t *testing.T) {
	var n noCopy

	n.Lock()
	defer n.Unlock()
}
