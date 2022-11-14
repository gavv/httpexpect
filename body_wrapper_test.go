package httpexpect

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBodyWrapperRewind(t *testing.T) {
	body := newMockBody("test_body")

	cancelled := false
	cancelFn := func() {
		cancelled = true
	}

	wr := newBodyWrapper(body, cancelFn)

	b, err := ioutil.ReadAll(wr)
	assert.NoError(t, err)
	assert.Equal(t, "test_body", string(b))

	assert.True(t, body.closed)
	assert.True(t, cancelled)

	err = wr.Close()
	assert.NoError(t, err)

	wr.Rewind()

	b, err = ioutil.ReadAll(wr)
	assert.NoError(t, err)
	assert.Equal(t, "test_body", string(b))
}

func TestBodyWrapperGetBody(t *testing.T) {
	body := newMockBody("test_body")

	wr := newBodyWrapper(body, nil)

	rd1, err := wr.GetBody()
	assert.NoError(t, err)

	rd2, err := wr.GetBody()
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(rd1)
	assert.NoError(t, err)
	assert.Equal(t, "test_body", string(b))

	b, err = ioutil.ReadAll(rd2)
	assert.NoError(t, err)
	assert.Equal(t, "test_body", string(b))
}

func TestBodyWrapperClose(t *testing.T) {
	body := newMockBody("test_body")

	cancelled := false
	cancelFn := func() {
		cancelled = true
	}

	wr := newBodyWrapper(body, cancelFn)

	err := wr.Close()
	assert.NoError(t, err)

	assert.True(t, body.closed)
	assert.True(t, cancelled)
}
