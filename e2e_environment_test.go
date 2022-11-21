package httpexpect

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestE2EEnvironmentDefault(t *testing.T) {
	e1 := Default(t, "")
	e2 := WithConfig(Config{
		Reporter: t,
	})

	assert.NotNil(t, e1.Env())
	assert.NotNil(t, e2.Env())

	assert.False(t, e1.Env().Has("key"))
	assert.False(t, e2.Env().Has("key"))

	e1.Env().Put("key", "value")

	assert.True(t, e1.Env().Has("key"))
	assert.False(t, e2.Env().Has("key"))

	e2.Env().Put("key", "value")

	assert.True(t, e1.Env().Has("key"))
	assert.True(t, e2.Env().Has("key"))
}

func TestE2EEnvironmentShared(t *testing.T) {
	env := NewEnvironment(t)

	e1 := WithConfig(Config{
		Reporter:    t,
		Environment: env,
	})
	e2 := WithConfig(Config{
		Reporter:    t,
		Environment: env,
	})

	assert.NotNil(t, e1.Env())
	assert.NotNil(t, e2.Env())

	assert.False(t, e1.Env().Has("key"))
	assert.False(t, e2.Env().Has("key"))

	e1.Env().Put("key", "value")

	assert.True(t, e1.Env().Has("key"))
	assert.True(t, e2.Env().Has("key"))
}
