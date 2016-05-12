package httpexpect

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestChainFail(t *testing.T) {
	chain := makeChain(mockReporter{t})

	assert.False(t, chain.failed())

	chain.fail("fail")
	assert.True(t, chain.failed())

	chain.fail("fail")
	assert.True(t, chain.failed())
}

func TestChainCopy(t *testing.T) {
	chain1 := makeChain(mockReporter{t})
	chain2 := chain1

	assert.False(t, chain1.failed())
	assert.False(t, chain2.failed())

	chain1.fail("fail")

	assert.True(t, chain1.failed())
	assert.False(t, chain2.failed())

	chain2.fail("fail")

	assert.True(t, chain1.failed())
	assert.True(t, chain2.failed())
}
