package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWebsocketCodesMessageType(t *testing.T) {
	for n := 0; n < 100; n++ {
		assert.NotEmpty(t, wsMessageType(n).String())
	}
}

func TestWebsocketCodesCloseCode(t *testing.T) {
	for n := 0; n < 2000; n++ {
		assert.NotEmpty(t, wsCloseCode(n).String())
	}
}
