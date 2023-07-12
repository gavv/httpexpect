package httpexpect

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBodyWrapper_Close(t *testing.T) {
	t.Run("close after read", func(t *testing.T) {
		body := newMockBody("test_body")

		cancelCount := 0
		cancelFn := func() {
			cancelCount++
		}

		wrp := newBodyWrapper(body, cancelFn)

		b, err := io.ReadAll(wrp)
		assert.NoError(t, err)
		assert.Equal(t, "test_body", string(b))

		// fully read and closed
		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, cancelCount)

		err = wrp.Close()
		assert.NoError(t, err)

		// nothing changed
		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, cancelCount)
	})

	t.Run("close before read", func(t *testing.T) {
		body := newMockBody("test_body")

		cancelCount := 0
		cancelFn := func() {
			cancelCount++
		}

		wrp := newBodyWrapper(body, cancelFn)

		err := wrp.Close()
		assert.NoError(t, err)

		// fully read and closed
		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, cancelCount)

		b := make([]byte, 1)

		n, err := wrp.Read(b)
		assert.Error(t, io.EOF, err)
		assert.Equal(t, 0, n)

		// nothing changed
		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, cancelCount)
	})
}

func TestBodyWrapper_Rewind(t *testing.T) {
	t.Run("readall - close - rewind - readall", func(t *testing.T) {
		body := newMockBody("test_body")

		cancelCount := 0
		cancelFn := func() {
			cancelCount++
		}

		wrp := newBodyWrapper(body, cancelFn)

		b, err := io.ReadAll(wrp)
		assert.NoError(t, err)
		assert.Equal(t, "test_body", string(b))

		readCount := body.readCount
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, cancelCount)

		err = wrp.Close()
		assert.NoError(t, err)

		wrp.Rewind()

		b, err = io.ReadAll(wrp)
		assert.NoError(t, err)
		assert.Equal(t, "test_body", string(b))

		assert.Equal(t, readCount, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, cancelCount)
	})

	t.Run("rewind - readall - close - rewind - readall", func(t *testing.T) {
		body := newMockBody("test_body")

		cancelCount := 0
		cancelFn := func() {
			cancelCount++
		}

		wrp := newBodyWrapper(body, cancelFn)

		wrp.Rewind()

		b, err := io.ReadAll(wrp)
		assert.NoError(t, err)
		assert.Equal(t, "test_body", string(b))

		readCount := body.readCount
		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, cancelCount)

		err = wrp.Close()
		assert.NoError(t, err)

		wrp.Rewind()

		b, err = io.ReadAll(wrp)
		assert.NoError(t, err)
		assert.Equal(t, "test_body", string(b))

		assert.Equal(t, readCount, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, cancelCount)
	})
}

func TestBodyWrapper_GetBody(t *testing.T) {
	t.Run("independent readers", func(t *testing.T) {
		body := newMockBody("test_body")

		wrp := newBodyWrapper(body, nil)

		rd1, err := wrp.GetBody()
		assert.NoError(t, err)

		rd2, err := wrp.GetBody()
		assert.NoError(t, err)

		b, err := io.ReadAll(rd1)
		assert.NoError(t, err)
		assert.Equal(t, "test_body", string(b))

		b, err = io.ReadAll(rd2)
		assert.NoError(t, err)
		assert.Equal(t, "test_body", string(b))

		assert.NotEqual(t, 0, body.readCount)
		assert.Equal(t, 1, body.closeCount)
	})
}

func TestBodyWrapper_OneError(t *testing.T) {
	bodyErr := errors.New("test_error")
	bodyText := "test_body"

	checkReadErr := func(t *testing.T, wrp *bodyWrapper) {
		b := make([]byte, len(bodyText)+1)
		n, err := wrp.Read(b)

		assert.Equal(t, bodyErr, err)
		assert.Equal(t, 0, n)
	}

	checkReadNoErr := func(t *testing.T, wrp *bodyWrapper) {
		b := make([]byte, len(bodyText)+1)
		n, err := wrp.Read(b)

		assert.NoError(t, err)
		assert.Equal(t, len(bodyText), n)
	}

	checkCloseErr := func(t *testing.T, wrp *bodyWrapper) {
		err := wrp.Close()

		assert.Equal(t, bodyErr, err)
	}

	checkCloseNoErr := func(t *testing.T, wrp *bodyWrapper) {
		err := wrp.Close()

		assert.NoError(t, err)
	}

	checkGetBodyErr := func(t *testing.T, wrp *bodyWrapper) {
		rd, err := wrp.GetBody()

		assert.Equal(t, bodyErr, err)
		assert.Nil(t, rd)
	}

	t.Run("read_err, read_close", func(t *testing.T) {
		body := newMockBody(bodyText)
		body.readErr = bodyErr

		wrp := newBodyWrapper(body, nil)

		checkReadErr(t, wrp)
		checkCloseNoErr(t, wrp)

		checkReadErr(t, wrp)
		checkCloseNoErr(t, wrp)
	})

	t.Run("read_err, close_read", func(t *testing.T) {
		body := newMockBody(bodyText)
		body.readErr = bodyErr

		wrp := newBodyWrapper(body, nil)

		checkCloseErr(t, wrp)
		checkReadErr(t, wrp)

		checkCloseNoErr(t, wrp)
		checkReadErr(t, wrp)
	})

	t.Run("close_err, read_close", func(t *testing.T) {
		body := newMockBody(bodyText)
		body.closeErr = bodyErr

		wrp := newBodyWrapper(body, nil)

		checkReadNoErr(t, wrp)
		checkCloseErr(t, wrp)

		checkReadErr(t, wrp)
		checkCloseErr(t, wrp)
	})

	t.Run("close_err, close_read", func(t *testing.T) {
		body := newMockBody(bodyText)
		body.closeErr = bodyErr

		wrp := newBodyWrapper(body, nil)

		checkCloseErr(t, wrp)
		checkReadErr(t, wrp)

		checkCloseErr(t, wrp)
		checkReadErr(t, wrp)
	})

	t.Run("read_err, get_body", func(t *testing.T) {
		body := newMockBody(bodyText)
		body.readErr = bodyErr

		wrp := newBodyWrapper(body, nil)

		checkGetBodyErr(t, wrp)
		checkGetBodyErr(t, wrp)
	})

	t.Run("close_err, get_body", func(t *testing.T) {
		body := newMockBody(bodyText)
		body.closeErr = bodyErr

		wrp := newBodyWrapper(body, nil)

		checkGetBodyErr(t, wrp)
		checkGetBodyErr(t, wrp)
	})
}

func TestBodyWrapper_TwoErrors(t *testing.T) {
	t.Run("read_close", func(t *testing.T) {
		bodyText := "test_body"
		body := newMockBody(bodyText)

		body.readErr = errors.New("read_error")
		body.closeErr = errors.New("close_error")

		wrp := newBodyWrapper(body, nil)

		b := make([]byte, len(bodyText)+1)
		var err error

		for i := 0; i < 2; i++ {
			_, err = wrp.Read(b)
			assert.Equal(t, body.readErr, err)

			err = wrp.Close()
			assert.Equal(t, body.closeErr, err)
		}
	})

	t.Run("close_read", func(t *testing.T) {
		bodyText := "test_body"
		body := newMockBody(bodyText)

		body.readErr = errors.New("read_error")
		body.closeErr = errors.New("close_error")

		wrp := newBodyWrapper(body, nil)

		b := make([]byte, len(bodyText)+1)
		var err error

		for i := 0; i < 2; i++ {
			err = wrp.Close()
			assert.Equal(t, body.closeErr, err)

			_, err = wrp.Read(b)
			assert.Equal(t, body.readErr, err)
		}
	})
}

func TestBodyWrapper_RewindError(t *testing.T) {
	bodyText := "test_body"
	body := newMockBody(bodyText)

	body.readErr = errors.New("read_error")
	body.closeErr = errors.New("close_error")

	wrp := newBodyWrapper(body, nil)

	b := make([]byte, len(bodyText)+1)
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

func TestBodyWrapper_Memory(t *testing.T) {
	t.Run("unaligned read", func(t *testing.T) {
		bodyText := "123456789"
		bodyChunks := []string{"1234", "5678", "9"}

		body := newMockBody(bodyText)
		wrp := newBodyWrapper(body, nil)

		b := make([]byte, 4)
		var (
			err error
			n   int
		)

		for i := 0; i < 2; i++ {
			n, err = wrp.Read(b)
			assert.NoError(t, err)
			assert.Equal(t, 4, n)
			assert.Equal(t, bodyChunks[i], string(b))

			assert.False(t, wrp.isFullyRead)
			assert.Equal(t, bodyText[:(i+1)*4], string(wrp.memBytes))

			assert.Equal(t, i+1, body.readCount)
			assert.Equal(t, 0, body.closeCount)
			assert.Equal(t, 0, body.eofCount)
		}

		// last read is incomplete
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 1, n)
		assert.Equal(t, bodyChunks[2], string(b[:n]))

		assert.False(t, wrp.isFullyRead) // not fully read until the next read
		assert.Equal(t, bodyText, string(wrp.memBytes))

		assert.Equal(t, 3, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// read eof
		n, err = wrp.Read(b)
		assert.Error(t, io.EOF, err)
		assert.Equal(t, 0, n)

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, bodyText, string(wrp.memBytes))

		assert.Equal(t, 4, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)
	})

	t.Run("aligned read", func(t *testing.T) {
		bodyText := "12345678"
		bodyChunks := []string{"1234", "5678"}

		body := newMockBody(bodyText)
		wrp := newBodyWrapper(body, nil)

		b := make([]byte, 4)
		var (
			err error
			n   int
		)

		for i := 0; i < 2; i++ {
			n, err = wrp.Read(b)
			assert.NoError(t, err)
			assert.Equal(t, 4, n)
			assert.Equal(t, bodyChunks[i], string(b))

			assert.False(t, wrp.isFullyRead)
			assert.Equal(t, bodyText[:(i+1)*4], string(wrp.memBytes))

			assert.Equal(t, i+1, body.readCount)
			assert.Equal(t, 0, body.closeCount)
			assert.Equal(t, 0, body.eofCount)
		}

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, bodyText, string(wrp.memBytes))

		// read eof
		n, err = wrp.Read(b)
		assert.Error(t, io.EOF, err)
		assert.Equal(t, 0, n)

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, bodyText, string(wrp.memBytes))

		assert.Equal(t, 3, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)
	})

	t.Run("partial read", func(t *testing.T) {
		buffer := bytes.NewBufferString("")
		body := &mockBody{
			reader: buffer,
		}

		wrp := newBodyWrapper(body, nil)

		b := make([]byte, 4)
		var (
			err error
			n   int
		)

		// add bytes to response
		buffer.Write([]byte("1234"))

		// partial read
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "1234", string(b[:4]))

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, "1234", string(wrp.memBytes))

		assert.Equal(t, 1, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// add bytes to response
		buffer.Write([]byte("5678"))

		// another partial read
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "5678", string(b[:4]))

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// read eof
		n, err = wrp.Read(b)
		assert.Error(t, io.EOF, err)
		assert.Equal(t, 0, n)

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 3, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)
	})

	t.Run("rewind", func(t *testing.T) {
		body := newMockBody("123456789")
		wrp := newBodyWrapper(body, nil)

		b := make([]byte, 4)
		var (
			err error
			n   int
		)

		// first read
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "1234", string(b))

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, "1234", string(wrp.memBytes))

		assert.Equal(t, 1, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// rewind
		wrp.Rewind()

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "123456789", string(wrp.memBytes))

		assert.Equal(t, 3, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// new first read
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "1234", string(b))

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "123456789", string(wrp.memBytes))

		assert.Equal(t, 3, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)
	})

	t.Run("get body", func(t *testing.T) {
		body := newMockBody("123456789")
		wrp := newBodyWrapper(body, nil)

		b := make([]byte, 4)
		var (
			err    error
			n      int
			reader io.ReadCloser
		)

		// first read
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "1234", string(b))

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, "1234", string(wrp.memBytes))

		assert.Equal(t, 1, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// get body between reads
		reader, err = wrp.GetBody()
		assert.NoError(t, err)
		assert.NotNil(t, reader)

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "123456789", string(wrp.memBytes))

		assert.Equal(t, 3, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// check body
		c, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.Equal(t, "123456789", string(c))

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "123456789", string(wrp.memBytes))

		assert.Equal(t, 3, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// second read
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "5678", string(b))

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "123456789", string(wrp.memBytes))

		assert.Equal(t, 3, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)
	})
}

func TestBodyWrapper_DisableRewinds(t *testing.T) {
	t.Run("disable rewinds before first read", func(t *testing.T) {
		body := newMockBody("12345678")
		wrp := newBodyWrapper(body, nil)

		b := make([]byte, 4)
		var (
			err    error
			n      int
			reader io.ReadCloser
		)

		// disable rewinds
		wrp.DisableRewinds()

		assert.False(t, wrp.isFullyRead)
		assert.Nil(t, wrp.memBytes)

		assert.Equal(t, 0, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// first read
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "1234", string(b))

		assert.False(t, wrp.isFullyRead)
		assert.Nil(t, wrp.memBytes)

		assert.Equal(t, 1, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// does not work
		reader, err = wrp.GetBody()
		assert.EqualError(t, err, "rewinds are disabled, cannot get body")
		assert.Nil(t, reader)

		// no-op
		wrp.Rewind()

		assert.False(t, wrp.isFullyRead)
		assert.Nil(t, wrp.memBytes)

		assert.Equal(t, 1, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// second read
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "5678", string(b))

		assert.False(t, wrp.isFullyRead)
		assert.Nil(t, wrp.memBytes)

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// read eof
		n, err = wrp.Read(b)
		assert.Error(t, io.EOF, err)
		assert.Equal(t, 0, n)

		assert.False(t, wrp.isFullyRead)
		assert.Nil(t, wrp.memBytes)

		assert.Equal(t, 3, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// close
		err = wrp.Close()
		assert.NoError(t, err)

		assert.False(t, wrp.isFullyRead)
		assert.Nil(t, wrp.memBytes)

		assert.Equal(t, 3, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)
	})

	t.Run("disable rewinds during read from http", func(t *testing.T) {
		body := newMockBody("12345678")
		wrp := newBodyWrapper(body, nil)

		b := make([]byte, 4)
		var (
			err    error
			n      int
			reader io.ReadCloser
		)

		// first read
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "1234", string(b))

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, "1234", string(wrp.memBytes))

		assert.Equal(t, 1, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// disable rewinds
		wrp.DisableRewinds()

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, "1234", string(wrp.memBytes))

		assert.Equal(t, 1, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// does not work
		reader, err = wrp.GetBody()
		assert.EqualError(t, err, "rewinds are disabled, cannot get body")
		assert.Nil(t, reader)

		// no-op
		wrp.Rewind()

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, "1234", string(wrp.memBytes))

		assert.Equal(t, 1, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// second read
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "5678", string(b))

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, "1234", string(wrp.memBytes))

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// read eof
		n, err = wrp.Read(b)
		assert.Error(t, io.EOF, err)
		assert.Equal(t, 0, n)

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, "1234", string(wrp.memBytes))

		assert.Equal(t, 3, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// close
		err = wrp.Close()
		assert.NoError(t, err)

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, "1234", string(wrp.memBytes))

		assert.Equal(t, 3, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)
	})

	t.Run("disable rewinds after full read from http", func(t *testing.T) {
		body := newMockBody("12345678")
		wrp := newBodyWrapper(body, nil)

		b := make([]byte, 8)
		var (
			err    error
			n      int
			reader io.ReadCloser
		)

		// first read
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 8, n)
		assert.Equal(t, "12345678", string(b))

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 1, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// read eof
		n, err = wrp.Read(b)
		assert.Error(t, io.EOF, err)
		assert.Equal(t, 0, n)

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// disable rewinds
		wrp.DisableRewinds()

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// does not work
		reader, err = wrp.GetBody()
		assert.EqualError(t, err, "rewinds are disabled, cannot get body")
		assert.Nil(t, reader)

		// no-op
		wrp.Rewind()

		// read eof
		n, err = wrp.Read(b)
		assert.Error(t, io.EOF, err)
		assert.Equal(t, 0, n)

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// close
		err = wrp.Close()
		assert.NoError(t, err)

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)
	})

	t.Run("disable rewinds during read from memory", func(t *testing.T) {
		body := newMockBody("12345678")
		wrp := newBodyWrapper(body, nil)

		b := make([]byte, 8)
		var (
			err    error
			n      int
			reader io.ReadCloser
		)

		// first read
		n, err = wrp.Read(b)
		assert.NoError(t, err)
		assert.Equal(t, 8, n)
		assert.Equal(t, "12345678", string(b))

		assert.False(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 1, body.readCount)
		assert.Equal(t, 0, body.closeCount)
		assert.Equal(t, 0, body.eofCount)

		// read eof
		n, err = wrp.Read(b)
		assert.Error(t, io.EOF, err)
		assert.Equal(t, 0, n)

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// rewind
		wrp.Rewind()

		// new first read
		n, err = wrp.Read(b[:4])
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "1234", string(b[:4]))

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// disable rewinds
		wrp.DisableRewinds()

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// does not work
		reader, err = wrp.GetBody()
		assert.EqualError(t, err, "rewinds are disabled, cannot get body")
		assert.Nil(t, reader)

		// no-op
		wrp.Rewind()

		// second read
		n, err = wrp.Read(b[:4])
		assert.NoError(t, err)
		assert.Equal(t, 4, n)
		assert.Equal(t, "5678", string(b[:4]))

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// read eof
		n, err = wrp.Read(b)
		assert.Error(t, io.EOF, err)
		assert.Equal(t, 0, n)

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)

		// no-op
		wrp.Rewind()

		// read eof
		n, err = wrp.Read(b)
		assert.Error(t, io.EOF, err)
		assert.Equal(t, 0, n)

		assert.True(t, wrp.isFullyRead)
		assert.Equal(t, "12345678", string(wrp.memBytes))

		assert.Equal(t, 2, body.readCount)
		assert.Equal(t, 1, body.closeCount)
		assert.Equal(t, 1, body.eofCount)
	})
}
