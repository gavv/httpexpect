package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainFail(t *testing.T) {
	chain := newMockChain(t)

	assert.False(t, chain.failed())

	chain.fail(&AssertionFailure{})
	assert.True(t, chain.failed())

	chain.fail(&AssertionFailure{})
	assert.True(t, chain.failed())
}

func TestChainClone(t *testing.T) {
	chain1 := newMockChain(t)
	chain2 := chain1.clone()

	assert.False(t, chain1.failed())
	assert.False(t, chain2.failed())

	chain1.fail(&AssertionFailure{})

	assert.True(t, chain1.failed())
	assert.False(t, chain2.failed())

	chain2.fail(&AssertionFailure{})

	assert.True(t, chain1.failed())
	assert.True(t, chain2.failed())
}

func TestChainReport(t *testing.T) {
	r0 := newMockReporter(t)

	chain := newDefaultChain("", r0)

	r1 := newMockReporter(t)

	chain.assertOK(r1)
	assert.False(t, r1.reported)

	chain.assertFailed(r1)
	assert.True(t, r1.reported)

	assert.False(t, chain.failed())

	chain.fail(&AssertionFailure{})
	assert.True(t, r0.reported)

	r2 := newMockReporter(t)

	chain.assertFailed(r2)
	assert.False(t, r2.reported)

	chain.assertOK(r2)
	assert.True(t, r2.reported)
}
