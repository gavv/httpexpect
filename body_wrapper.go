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
//  - use Read to read body contents and Rewind to restart reading from beginning
//  - use GetBody to get new reader for body contents
type bodyWrapper struct {
	currReader io.Reader

	origReader io.ReadCloser
	origBytes  []byte
	origErr    error

	cancelFunc context.CancelFunc

	isInitialized bool

	mu sync.Mutex
}

func newBodyWrapper(reader io.ReadCloser, cancelFunc context.CancelFunc) *bodyWrapper {
	bw := &bodyWrapper{
		origReader: reader,
		cancelFunc: cancelFunc,
	}

	// This is not strictly necessary because we should always call close.
	// This is just a reinsurance.
	runtime.SetFinalizer(bw, (*bodyWrapper).Close)

	return bw
}

// Read body contents
func (bw *bodyWrapper) Read(p []byte) (n int, err error) {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	// Preserve original reader error
	if bw.origErr != nil {
		return 0, bw.origErr
	}

	// Lazy initialization
	if !bw.isInitialized {
		if err := bw.initialize(); err != nil {
			return 0, err
		}
	}

	if bw.currReader == nil {
		bw.currReader = bytes.NewReader(bw.origBytes)
	}
	return bw.currReader.Read(p)
}

// Close body
func (bw *bodyWrapper) Close() error {
	bw.mu.Lock()
	defer bw.mu.Unlock()

	err := bw.origErr

	// Rewind or GetBody may be called later, so be sure to
	// read body into memory before closing
	if !bw.isInitialized {
		initErr := bw.initialize()
		if initErr != nil {
			err = initErr
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
	if !bw.isInitialized {
		return
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
	if bw.origErr != nil {
		return nil, bw.origErr
	}

	// Lazy initialization
	if !bw.isInitialized {
		if err := bw.initialize(); err != nil {
			return nil, err
		}
	}

	return ioutil.NopCloser(bytes.NewReader(bw.origBytes)), nil
}

func (bw *bodyWrapper) initialize() error {
	if !bw.isInitialized {
		bw.isInitialized = true

		if bw.origReader != nil {
			bw.origBytes, bw.origErr = ioutil.ReadAll(bw.origReader)

			_ = bw.closeAndCancel()

			if bw.origErr != nil {
				return bw.origErr
			}
		}
	}

	return nil
}

func (bw *bodyWrapper) closeAndCancel() error {
	if bw.origReader == nil && bw.cancelFunc == nil {
		return nil
	}

	var err error

	if bw.origReader != nil {
		err = bw.origReader.Close()
		bw.origReader = nil

		if bw.origErr == nil {
			bw.origErr = err
		}
	}

	if bw.cancelFunc != nil {
		bw.cancelFunc()
		bw.cancelFunc = nil
	}

	// Finalizer is not needed anymore.
	runtime.SetFinalizer(bw, nil)

	return err
}
