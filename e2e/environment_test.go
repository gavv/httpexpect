package e2e

import (
	"testing"

	"github.com/gavv/httpexpect/v2"
	"github.com/stretchr/testify/assert"
)

func TestE2EEnvironment_Default(t *testing.T) {
	e1 := httpexpect.Default(t, "")
	e2 := httpexpect.WithConfig(httpexpect.Config{
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

func TestE2EEnvironment_Shared(t *testing.T) {
	env := httpexpect.NewEnvironment(t)

	e1 := httpexpect.WithConfig(httpexpect.Config{
		Reporter:    t,
		Environment: env,
	})
	e2 := httpexpect.WithConfig(httpexpect.Config{
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
