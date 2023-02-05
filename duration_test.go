package httpexpect

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestDuration_FailedChain(t *testing.T) {
	chain := newMockChain(t)
	chain.setFailed()

	tm := time.Second
	value := newDuration(chain, &tm)
	value.chain.assertFailed(t)

	value.Alias("foo")
	value.IsEqual(tm)
	value.NotEqual(tm)
	value.InRange(tm, tm)
	value.NotInRange(tm, tm)
	value.InList(tm)
	value.NotInList(tm)
	value.Gt(tm)
	value.Ge(tm)
	value.Lt(tm)
	value.Le(tm)
}

func TestDuration_Constructors(t *testing.T) {
	tm := time.Second

	t.Run("reporter", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewDuration(reporter, tm)
		value.IsEqual(tm)
		value.chain.assertNotFailed(t)
	})

	t.Run("config", func(t *testing.T) {
		reporter := newMockReporter(t)
		value := NewDurationC(Config{
			Reporter: reporter,
		}, tm)
		value.IsEqual(tm)
		value.chain.assertNotFailed(t)
	})

	t.Run("chain", func(t *testing.T) {
		chain := newMockChain(t)
		value := newDuration(chain, &tm)
		assert.NotSame(t, value.chain, chain)
		assert.Equal(t, value.chain.context.Path, chain.context.Path)
	})

}

func TestDuration_Alias(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)
	assert.Equal(t, []string{"Duration()"}, value.chain.context.Path)
	assert.Equal(t, []string{"Duration()"}, value.chain.context.AliasedPath)

	value.Alias("foo")
	assert.Equal(t, []string{"Duration()"}, value.chain.context.Path)
	assert.Equal(t, []string{"foo"}, value.chain.context.AliasedPath)
}

func TestDuration_Set(t *testing.T) {
	chain := newMockChain(t)

	tm := time.Second
	value := newDuration(chain, &tm)

	value.IsSet()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotSet()
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDuration_Unset(t *testing.T) {
	chain := newMockChain(t)

	value := newDuration(chain, nil)

	value.IsSet()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotSet()
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestDuration_Equal(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	assert.Equal(t, time.Second, value.Raw())

	value.IsEqual(time.Second)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.IsEqual(time.Minute)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotEqual(time.Minute)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotEqual(time.Second)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDuration_Greater(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	value.Gt(time.Second - 1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Gt(time.Second)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Ge(time.Second - 1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Ge(time.Second)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Ge(time.Second + 1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDuration_Lesser(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	value.Lt(time.Second + 1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Lt(time.Second)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.Le(time.Second + 1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Le(time.Second)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.Le(time.Second - 1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()
}

func TestDuration_InRange(t *testing.T) {
	reporter := newMockReporter(t)

	value := NewDuration(reporter, time.Second)

	value.InRange(time.Second, time.Second)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Second, time.Second)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Second-1, time.Second)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Second-1, time.Second)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Second, time.Second+1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Second, time.Second+1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Second+1, time.Second+2)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Second+1, time.Second+2)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Second-2, time.Second-1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Second-2, time.Second-1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InRange(time.Second+1, time.Second-1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInRange(time.Second+1, time.Second-1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}

func TestDuration_InList(t *testing.T) {
	reporter := newMockReporter(t)

	newDuration(newMockChain(t), nil).InList(time.Second).chain.assertFailed(t)
	newDuration(newMockChain(t), nil).NotInList(time.Second).chain.assertFailed(t)

	value := NewDuration(reporter, time.Second)

	value.InList()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList()
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(time.Second, time.Minute)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInList(time.Second, time.Minute)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(time.Second-1, time.Minute)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList(time.Second-1, time.Minute)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.InList(time.Second, time.Second+1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()

	value.NotInList(time.Second, time.Second+1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.InList(time.Second+1, time.Second-1)
	value.chain.assertFailed(t)
	value.chain.clearFailed()

	value.NotInList(time.Second+1, time.Second-1)
	value.chain.assertNotFailed(t)
	value.chain.clearFailed()
}
