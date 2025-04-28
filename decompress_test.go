package httpexpect

import (
	"bytes"
	"compress/gzip"
	"testing"
)

// Helper function to compress a string (for testing purposes)
func compressString(input string) ([]byte, error) {
	var buf bytes.Buffer
	writer := gzip.NewWriter(&buf)

	_, err := writer.Write([]byte(input))
	if err != nil {
		return nil, err
	}

	err = writer.Close()
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func TestDecompressGZIP(t *testing.T) {
	originalText := "Hello, this is a test string for decompression."

	// First, compress the original text
	compressedData, err := compressString(originalText)
	if err != nil {
		t.Fatalf("Failed to compress string: %v", err)
	}

	// Now, decompress it using your function
	decompressedText, err := DecompressGZIP(compressedData)
	if err != nil {
		t.Fatalf("Failed to decompress string: %v", err)
	}

	// Verify
	if decompressedText != originalText {
		t.Errorf("Expected %q but got %q", originalText, decompressedText)
	}
}
