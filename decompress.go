package httpexpect

import (
	"bytes"
	"compress/gzip"
	"io/ioutil"
)

// DecompressGZIP decompresses a GZIP compressed string and returns the original string
func DecompressGZIP(compressedData []byte) (string, error) {
	reader, err := gzip.NewReader(bytes.NewBuffer(compressedData))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	decompressedData, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(decompressedData), nil
}
