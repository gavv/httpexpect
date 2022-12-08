package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainFail(t *testing.T) {
	chain := newMockChain(t)

	assert.False(t, chain.failed())

	chain.fail(AssertionFailure{})
	assert.True(t, chain.failed())

	chain.fail(AssertionFailure{})
	assert.True(t, chain.failed())
}

func TestChainClone(t *testing.T) {
	chain1 := newMockChain(t)
	chain2 := chain1.clone()

	assert.False(t, chain1.failed())
	assert.False(t, chain2.failed())

	chain1.fail(AssertionFailure{})

	assert.True(t, chain1.failed())
	assert.False(t, chain2.failed())

	chain2.fail(AssertionFailure{})

	assert.True(t, chain1.failed())
	assert.True(t, chain2.failed())
}

func TestChainReport(t *testing.T) {
	r0 := newMockReporter(t)

	chain := newChainWithDefaults("test", r0)

	r1 := newMockReporter(t)

	chain.assertOK(r1)
	assert.False(t, r1.reported)

	chain.assertFailed(r1)
	assert.True(t, r1.reported)

	assert.False(t, chain.failed())

	chain.fail(AssertionFailure{})
	assert.True(t, r0.reported)

	r2 := newMockReporter(t)

	chain.assertFailed(r2)
	assert.False(t, r2.reported)

	chain.assertOK(r2)
	assert.True(t, r2.reported)
}

func TestChainHandler(t *testing.T) {
	handler := &mockAssertionHandler{}

	chain := newChainWithConfig("test", Config{
		AssertionHandler: handler,
	})

	chain.enter("test")
	chain.fail(AssertionFailure{})
	chain.leave()

	assert.NotNil(t, handler.ctx)
	assert.NotNil(t, handler.failure)

	chain.reset()

	handler.ctx = nil
	handler.failure = nil

	chain.enter("test")
	chain.leave()

	assert.NotNil(t, handler.ctx)
	assert.Nil(t, handler.failure)
}

func TestChainSeverity(t *testing.T) {
	handler := &mockAssertionHandler{}

	chain := newChainWithConfig("test", Config{
		AssertionHandler: handler,
	})

	chain.fail(AssertionFailure{})

	assert.NotNil(t, handler.failure)
	assert.Equal(t, SeverityError, handler.failure.Severity)

	chain.reset()

	chain.setSeverity(SeverityError)
	chain.fail(AssertionFailure{})

	assert.NotNil(t, handler.failure)
	assert.Equal(t, SeverityError, handler.failure.Severity)

	chain.reset()

	chain.setSeverity(SeverityInfo)
	chain.fail(AssertionFailure{})

	assert.NotNil(t, handler.failure)
	assert.Equal(t, SeverityInfo, handler.failure.Severity)
}

func TestChainCallback(t *testing.T) {
	handler := &mockAssertionHandler{}

	chain := newChainWithConfig("test", Config{
		AssertionHandler: handler,
	})

	called := false

	chain.setFailCallback(func() {
		called = true
	})

	chain.fail(AssertionFailure{})
	assert.True(t, called)
}
