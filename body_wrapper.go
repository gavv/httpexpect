package httpexpect

import (
	"bytes"
	"context"
	"errors"
	"io"
	"io/ioutil"
	"runtime"
	"sync"
)

// Wrapper for request or response body reader.
// Allows to read body multiple times using two approaches:
//   - use Read to read body contents and Rewind to restart reading from beginning
//   - use GetBody to get new reader for body contents
type bodyWrapper struct {
	// Protects all operations.
	mu sync.Mutex

	// Original reader of HTTP response body.
	httpReader io.ReadCloser

	// Cancellation function for original HTTP response.
	// If set, called after HTTP response is fully read into memory.
	httpCancelFunc context.CancelFunc

	// Reader for HTTP response body stored in memory.
	// Rewind() resets this reader to start from the beginning.
	memReader io.Reader

	// HTTP response body stored in memory.
	memBytes []byte

	// Cached read and close errors.
	readErr  error
	closeErr error

	// If true, Read will not store bytes in memory, and memBytes and memReader
	// won't be used.
	isRewindDisabled bool

	// True means that HTTP response was fully read into memory already.
	isFullyRead bool

	// True means that a read operation of any type was called at least once.
	isReadBefore bool
}

func newBodyWrapper(reader io.ReadCloser, cancelFunc context.CancelFunc) *bodyWrapper {
	bw := &bodyWrapper{
		httpReader:     reader,
		httpCancelFunc: cancelFunc,
	}

	// Finalizer will close body if closeAndCancel was never called.
	runtime.SetFinalizer(bw, (*bodyWrapper).Close)

	return bw
}

// Read body contents.
func (bw *bodyWrapper) Read(p []byte) (int, error) {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	bw.isReadBefore = true

	if bw.isRewindDisabled && !bw.isFullyRead {
		// Regular read from original HTTP response.
		if bw.readErr != nil {
			return 0, bw.readErr
		}
		return bw.httpReader.Read(p)
	} else if !bw.isFullyRead {
		// Read from original HTTP response + store into memory.
		if bw.readErr != nil {
			return 0, bw.readErr
		}
		return bw.httpReadNext(p)
	} else {
		// Read from memory.
		n, err := bw.memReader.Read(p)
		if err == io.EOF && bw.readErr != nil {
			err = bw.readErr
		}
		return n, err
	}
}

// Close body.
func (bw *bodyWrapper) Close() error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	// Preserve original reader error.
	err := bw.closeErr

	// Rewind or GetBody may be called later, so be sure to
	// read body into memory before closing.
	if !bw.isRewindDisabled && !bw.isFullyRead {
		bw.isReadBefore = true

		if readErr := bw.httpReadFull(); readErr != nil {
			err = readErr
		}
	}

	// Close original reader.
	closeErr := bw.closeAndCancel()
	if closeErr != nil {
		err = closeErr
	}

	// Reset memory reader.
	bw.memReader = bytes.NewReader(nil)

	return err
}

// Rewind reading to the beginning.
func (bw *bodyWrapper) Rewind() {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	// Rewind is no-op if disabled.
	if bw.isRewindDisabled {
		return
	}

	// Rewind is no-op until first read operation.
	if !bw.isReadBefore {
		return
	}

	// If HTTP response is not fully read yet, do it now.
	// If error occurs, it will be reported next read operation.
	if !bw.isFullyRead {
		_ = bw.httpReadFull()
	}

	// Reset reader to the beginning of memory chunk.
	bw.memReader = bytes.NewReader(bw.memBytes)
}

// Create new reader to retrieve body contents.
// New reader always reads body from the beginning.
// Does not affected by Rewind().
func (bw *bodyWrapper) GetBody() (io.ReadCloser, error) {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	bw.isReadBefore = true

	// Preserve original reader error.
	if bw.readErr != nil {
		return nil, bw.readErr
	}

	// GetBody() requires rewinds to be enabled.
	if bw.isRewindDisabled {
		return nil, errors.New("rewinds are disabled, cannot get body")
	}

	// If HTTP response is not fully read yet, do it now.
	if !bw.isFullyRead {
		if err := bw.httpReadFull(); err != nil {
			return nil, err
		}
	}

	// Return fresh reader for memory chunk.
	return ioutil.NopCloser(bytes.NewReader(bw.memBytes)), nil
}

// Disables storing body contents in memory and clears the cache.
func (bw *bodyWrapper) DisableRewinds() error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	bw.isRewindDisabled = true

	bw.memReader = nil
	bw.memBytes = nil

	return nil
}

func (bw *bodyWrapper) httpReadNext(p []byte) (n int, err error) {
	n, err = bw.httpReader.Read(p)

	if n > 0 {
		bw.memBytes = append(bw.memBytes, p[:n]...)
	}

	if err != nil {
		if err != io.EOF {
			bw.readErr = err
		}
		if closeErr := bw.closeAndCancel(); closeErr != nil && err == io.EOF {
			err = closeErr
		}

		bw.isFullyRead = true
		bw.memReader = bytes.NewReader(nil)
	}

	return
}

func (bw *bodyWrapper) httpReadFull() error {
	b, err := ioutil.ReadAll(bw.httpReader)

	bw.isFullyRead = true
	bw.memBytes = append(bw.memBytes, b...)
	bw.memReader = bytes.NewReader(bw.memBytes[len(bw.memBytes)-len(b):])

	if err != nil {
		bw.readErr = err
	}

	if closeErr := bw.closeAndCancel(); closeErr != nil && err == nil {
		err = closeErr
	}

	return err
}

func (bw *bodyWrapper) closeAndCancel() error {
	if bw.httpReader == nil && bw.httpCancelFunc == nil {
		return bw.closeErr
	}

	if bw.httpReader != nil {
		err := bw.httpReader.Close()
		bw.httpReader = nil

		if bw.readErr == nil {
			bw.readErr = err
		}

		if bw.closeErr == nil {
			bw.closeErr = err
		}
	}

	if bw.httpCancelFunc != nil {
		bw.httpCancelFunc()
		bw.httpCancelFunc = nil
	}

	// Finalizer is not needed anymore.
	runtime.SetFinalizer(bw, nil)

	return bw.closeErr
}
