package httpexpect

import (
	"errors"
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBodyWrapper_Rewind(t *testing.T) {
	body := newMockBody("test_body")

	cancelCount := 0
	cancelFn := func() {
		cancelCount++
	}

	wrp := newBodyWrapper(body, cancelFn)

	b, err := ioutil.ReadAll(wrp)
	assert.NoError(t, err)
	assert.Equal(t, "test_body", string(b))

	readCount := body.readCount
	assert.NotEqual(t, 0, body.readCount)
	assert.Equal(t, 1, body.closeCount)
	assert.Equal(t, 1, cancelCount)

	err = wrp.Close()
	assert.NoError(t, err)

	wrp.Rewind()

	b, err = ioutil.ReadAll(wrp)
	assert.NoError(t, err)
	assert.Equal(t, "test_body", string(b))

	assert.Equal(t, readCount, body.readCount)
	assert.Equal(t, 1, body.closeCount)
	assert.Equal(t, 1, cancelCount)
}

func TestBodyWrapper_GetBody(t *testing.T) {
	body := newMockBody("test_body")

	wrp := newBodyWrapper(body, nil)

	rd1, err := wrp.GetBody()
	assert.NoError(t, err)

	rd2, err := wrp.GetBody()
	assert.NoError(t, err)

	b, err := ioutil.ReadAll(rd1)
	assert.NoError(t, err)
	assert.Equal(t, "test_body", string(b))

	b, err = ioutil.ReadAll(rd2)
	assert.NoError(t, err)
	assert.Equal(t, "test_body", string(b))

	assert.NotEqual(t, 0, body.readCount)
	assert.Equal(t, 1, body.closeCount)
}

func TestBodyWrapper_Close(t *testing.T) {
	body := newMockBody("test_body")

	cancelCount := 0
	cancelFn := func() {
		cancelCount++
	}

	wrp := newBodyWrapper(body, cancelFn)

	err := wrp.Close()
	assert.NoError(t, err)

	assert.NotEqual(t, 0, body.readCount)
	assert.Equal(t, 1, body.closeCount)
	assert.Equal(t, 1, cancelCount)
}

func TestBodyWrapper_OneError(t *testing.T) {
	bodyErr := errors.New("test_error")

	checkReadErr := func(t *testing.T, wrp *bodyWrapper) {
		b := make([]byte, 10)
		n, err := wrp.Read(b)

		assert.Equal(t, bodyErr, err)
		assert.Equal(t, 0, n)
	}

	checkCloseErr := func(t *testing.T, wrp *bodyWrapper) {
		err := wrp.Close()

		assert.Equal(t, bodyErr, err)
	}

	checkCloseNoErr := func(t *testing.T, wrp *bodyWrapper) {
		err := wrp.Close()

		assert.Nil(t, err)
	}

	checkGetBodyErr := func(t *testing.T, wrp *bodyWrapper) {
		rd, err := wrp.GetBody()

		assert.Equal(t, bodyErr, err)
		assert.Nil(t, rd)
	}

	t.Run("readerr/read_close", func(t *testing.T) {
		body := newMockBody("test_body")
		body.readErr = bodyErr

		wrp := newBodyWrapper(body, nil)

		checkReadErr(t, wrp)
		checkCloseNoErr(t, wrp)

		checkReadErr(t, wrp)
		checkCloseNoErr(t, wrp)
	})

	t.Run("readerr/close_read", func(t *testing.T) {
		body := newMockBody("test_body")
		body.readErr = bodyErr

		wrp := newBodyWrapper(body, nil)

		checkCloseErr(t, wrp)
		checkReadErr(t, wrp)

		checkCloseNoErr(t, wrp)
		checkReadErr(t, wrp)
	})

	t.Run("closeerr/read_close", func(t *testing.T) {
		body := newMockBody("test_body")
		body.closeErr = bodyErr

		wrp := newBodyWrapper(body, nil)

		checkReadErr(t, wrp)
		checkCloseErr(t, wrp)

		checkReadErr(t, wrp)
		checkCloseErr(t, wrp)
	})

	t.Run("closeerr/close_read", func(t *testing.T) {
		body := newMockBody("test_body")
		body.closeErr = bodyErr

		wrp := newBodyWrapper(body, nil)

		checkCloseErr(t, wrp)
		checkReadErr(t, wrp)

		checkCloseErr(t, wrp)
		checkReadErr(t, wrp)
	})

	t.Run("readerr/getbody", func(t *testing.T) {
		body := newMockBody("test_body")
		body.readErr = bodyErr

		wrp := newBodyWrapper(body, nil)

		checkGetBodyErr(t, wrp)
		checkGetBodyErr(t, wrp)
	})

	t.Run("closeerr/getbody", func(t *testing.T) {
		body := newMockBody("test_body")
		body.closeErr = bodyErr

		wrp := newBodyWrapper(body, nil)

		checkGetBodyErr(t, wrp)
		checkGetBodyErr(t, wrp)
	})
}

func TestBodyWrapper_TwoErrors(t *testing.T) {
	t.Run("read_close", func(t *testing.T) {
		body := newMockBody("test_body")

		body.readErr = errors.New("read_error")
		body.closeErr = errors.New("close_error")

		wrp := newBodyWrapper(body, nil)

		b := make([]byte, 10)
		var err error

		for i := 0; i < 2; i++ {
			_, err = wrp.Read(b)
			assert.Equal(t, body.readErr, err)

			err = wrp.Close()
			assert.Equal(t, body.closeErr, err)
		}
	})

	t.Run("close_read", func(t *testing.T) {
		body := newMockBody("test_body")

		body.readErr = errors.New("read_error")
		body.closeErr = errors.New("close_error")

		wrp := newBodyWrapper(body, nil)

		b := make([]byte, 10)
		var err error

		for i := 0; i < 2; i++ {
			err = wrp.Close()
			assert.Equal(t, body.closeErr, err)

			_, err = wrp.Read(b)
			assert.Equal(t, body.readErr, err)
		}
	})
}

func TestBodyWrapper_ErrorRewind(t *testing.T) {
	body := newMockBody("test_body")

	body.readErr = errors.New("read_error")
	body.closeErr = errors.New("close_error")

	wrp := newBodyWrapper(body, nil)

	b := make([]byte, 10)
	var err error

	for i := 0; i < 2; i++ {
		_, err = wrp.Read(b)
		assert.NotNil(t, err)

		err = wrp.Close()
		assert.NotNil(t, err)

		_, err = wrp.GetBody()
		assert.NotNil(t, err)
	}

	wrp.Rewind()

	body.readErr = nil
	body.closeErr = nil

	for i := 0; i < 2; i++ {
		_, err = wrp.Read(b)
		assert.NotNil(t, err)

		err = wrp.Close()
		assert.NotNil(t, err)

		_, err = wrp.GetBody()
		assert.NotNil(t, err)
	}
}
