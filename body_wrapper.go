package httpexpect

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"runtime"
	"sync"
)

// Wrapper for request or response body reader
// Allows to read body multiple times using two approaches:
//   - use Read to read body contents and Rewind to restart reading from beginning
//   - use GetBody to get new reader for body contents
type bodyWrapper struct {
	currReader io.Reader

	origReader io.ReadCloser
	origBytes  []byte

	readErr  error
	closeErr error

	cancelFunc context.CancelFunc

	isFullyRead bool

	mu sync.Mutex
}

func newBodyWrapper(reader io.ReadCloser, cancelFunc context.CancelFunc) *bodyWrapper {
	bw := &bodyWrapper{
		origReader: reader,
		currReader: reader,
		cancelFunc: cancelFunc,
	}

	// Finalizer will close body if closeAndCancel was never called.
	runtime.SetFinalizer(bw, (*bodyWrapper).Close)

	return bw
}

// Read body contents
func (bw *bodyWrapper) Read(p []byte) (n int, err error) {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	// Preserve original reader error
	if bw.readErr != nil {
		return 0, bw.readErr
	}

	// Cache bytes in memory
	if !bw.isFullyRead {
		n, err = bw.origReader.Read(p)
		bw.origBytes = append(bw.origBytes, p[:n]...)
	} else {
		n, err = bw.currReader.Read(p)
	}

	if err != nil || n < len(p) {
		bw.isFullyRead = true // prevent further reads
		if err != nil && err != io.EOF {
			bw.readErr = err
		}
		if closeErr := bw.closeAndCancel(); closeErr != nil && (err == nil || err == io.EOF) {
			err = closeErr
			n = 0
		}
	}

	return
}

// Close body
func (bw *bodyWrapper) Close() error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	err := bw.closeErr

	// Rewind or GetBody may be called later, so be sure to
	// read body into memory before closing
	if !bw.isFullyRead {
		if readFullErr := bw.readFull(); readFullErr != nil {
			err = readFullErr
		}
	}

	// Close original reader
	closeErr := bw.closeAndCancel()
	if closeErr != nil {
		err = closeErr
	}

	return err
}

// Rewind reading to the beginning
func (bw *bodyWrapper) Rewind() {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	// Until first read, rewind is no-op
	if !bw.isFullyRead {
		_ = bw.readFull()
	}

	// Reset reader
	bw.currReader = bytes.NewReader(bw.origBytes)
}

// Create new reader to retrieve body contents
// New reader always reads body from the beginning
// Does not affected by Rewind()
func (bw *bodyWrapper) GetBody() (io.ReadCloser, error) {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	// Preserve original reader error
	if bw.readErr != nil {
		return nil, bw.readErr
	}

	if !bw.isFullyRead {
		if err := bw.readFull(); err != nil {
			return nil, err
		}
	}

	return ioutil.NopCloser(bytes.NewReader(bw.origBytes)), nil
}

// Reads the body fully, then cancels and closes the reader
func (bw *bodyWrapper) readFull() error {
	remainingBytes, err := ioutil.ReadAll(bw.currReader)
	if err != nil {
		bw.readErr = err
		bw.isFullyRead = true // Prevent further reads
		bw.closeAndCancel()
		return err
	}
	bw.origBytes = append(bw.origBytes, remainingBytes...)
	bw.isFullyRead = true
	return bw.closeAndCancel()
}

func (bw *bodyWrapper) closeAndCancel() error {
	if bw.origReader == nil && bw.cancelFunc == nil {
		return bw.closeErr
	}

	if bw.origReader != nil {
		err := bw.origReader.Close()
		bw.origReader = nil

		if bw.readErr == nil {
			bw.readErr = err
		}

		if bw.closeErr == nil {
			bw.closeErr = err
		}
	}

	if bw.cancelFunc != nil {
		bw.cancelFunc()
		bw.cancelFunc = nil
	}

	// Finalizer is not needed anymore.
	runtime.SetFinalizer(bw, nil)

	return bw.closeErr
}
