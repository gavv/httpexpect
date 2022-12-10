package httpexpect

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestChainFail(t *testing.T) {
	chain := newMockChain(t)

	assert.False(t, chain.failed())

	chain.fail(mockFailure())
	assert.True(t, chain.failed())

	chain.fail(mockFailure())
	assert.True(t, chain.failed())
}

func TestChainClone(t *testing.T) {
	chain1 := newMockChain(t)
	chain2 := chain1.clone()

	assert.False(t, chain1.failed())
	assert.False(t, chain2.failed())

	chain1.fail(mockFailure())

	assert.True(t, chain1.failed())
	assert.False(t, chain2.failed())

	chain2.fail(mockFailure())

	assert.True(t, chain1.failed())
	assert.True(t, chain2.failed())
}

func TestChainEnv(t *testing.T) {
	t.Run("newChainWithConfig", func(t *testing.T) {
		env1 := NewEnvironment(newMockReporter(t))
		chain1 := newChainWithConfig("root", Config{
			Environment: env1,
		})
		assert.True(t, env1 == chain1.getEnv())

		chain2 := newChainWithConfig("root", Config{
			Environment: nil,
		})
		assert.NotNil(t, chain2.getEnv())
	})

	t.Run("newChainWithDefaults", func(t *testing.T) {
		chain := newChainWithDefaults("root", newMockReporter(t))
		assert.NotNil(t, chain.getEnv())
	})
}

func TestChainRoot(t *testing.T) {
	t.Run("newChainWithConfig", func(t *testing.T) {
		chain1 := newChainWithConfig("root", Config{
			AssertionHandler: &mockAssertionHandler{},
		})
		assert.Equal(t, []string{"root"}, chain1.context.Path)

		chain2 := newChainWithConfig("", Config{
			AssertionHandler: &mockAssertionHandler{},
		})
		assert.Equal(t, []string{}, chain2.context.Path)
	})

	t.Run("newChainWithDefaults", func(t *testing.T) {
		chain1 := newChainWithDefaults("root", newMockReporter(t))
		assert.Equal(t, []string{"root"}, chain1.context.Path)

		chain2 := newChainWithDefaults("", newMockReporter(t))
		assert.Equal(t, []string{}, chain2.context.Path)
	})
}

func TestChainPath(t *testing.T) {
	path := func(c *chain) string {
		return strings.Join(c.context.Path, ".")
	}

	chain := newChainWithDefaults("root", newMockReporter(t))

	assert.Equal(t, "root", path(chain))

	chain.enter("foo")
	assert.Equal(t, "root.foo", path(chain))

	chain.enter("bar")
	assert.Equal(t, "root.foo.bar", path(chain))

	chainClone := chain.clone()
	chainClone.enter("baz")

	assert.Equal(t, "root.foo.bar", path(chain))
	assert.Equal(t, "root.foo.bar.baz", path(chainClone))

	chain.replace("qux")
	chainClone.replace("wee")

	assert.Equal(t, "root.foo.qux", path(chain))
	assert.Equal(t, "root.foo.bar.wee", path(chainClone))

	chain.leave()
	chainClone.leave()

	assert.Equal(t, "root.foo", path(chain))
	assert.Equal(t, "root.foo.bar", path(chainClone))

	chain.leave()
	chainClone.leave()

	assert.Equal(t, "root", path(chain))
	assert.Equal(t, "root.foo", path(chainClone))

	chain.leave()
	chainClone.leave()

	assert.Equal(t, "", path(chain))
	assert.Equal(t, "root", path(chainClone))
}

func TestChainPanic(t *testing.T) {
	t.Run("unpaired leave", func(t *testing.T) {
		chain := newChainWithDefaults("", newMockReporter(t))

		assert.Panics(t, func() {
			chain.leave()
		})
	})

	t.Run("unpaired enter/leave", func(t *testing.T) {
		chain := newChainWithDefaults("", newMockReporter(t))

		chain.enter("foo")
		chain.leave()

		assert.Panics(t, func() {
			chain.leave()
		})
	})

	t.Run("unpaired replace", func(t *testing.T) {
		chain := newChainWithDefaults("", newMockReporter(t))

		assert.Panics(t, func() {
			chain.replace("foo")
		})
	})

	t.Run("unpaired enter/replace", func(t *testing.T) {
		chain := newChainWithDefaults("", newMockReporter(t))

		chain.enter("foo")
		chain.leave()

		assert.Panics(t, func() {
			chain.replace("bar")
		})
	})
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

	chain.fail(mockFailure())
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
	chain.fail(mockFailure())
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

	chain.fail(mockFailure())

	assert.NotNil(t, handler.failure)
	assert.Equal(t, SeverityError, handler.failure.Severity)

	chain.reset()

	chain.setSeverity(SeverityError)
	chain.fail(mockFailure())

	assert.NotNil(t, handler.failure)
	assert.Equal(t, SeverityError, handler.failure.Severity)

	chain.reset()

	chain.setSeverity(SeverityLog)
	chain.fail(mockFailure())

	assert.NotNil(t, handler.failure)
	assert.Equal(t, SeverityLog, handler.failure.Severity)
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

	chain.fail(mockFailure())
	assert.True(t, called)
}
