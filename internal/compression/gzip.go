package compression

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"strings"
)

func Compress(data []byte) ([]byte, error) {

	var valByte bytes.Buffer
	writer := gzip.NewWriter(&valByte)
	if _, err := writer.Write(data); err != nil {
		return nil, fmt.Errorf("failed write data to compress temporary buffer: %v", err)
	}
	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("failed compress data: %v", err)
	}
	return valByte.Bytes(), nil
}

func Decompress(data []byte) ([]byte, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("failed init decompress reader: %v", err)
	}
	defer reader.Close()

	var valByte bytes.Buffer
	if _, err := valByte.ReadFrom(reader); err != nil {
		return nil, fmt.Errorf("failed decompress data: %v", err)
	}

	return valByte.Bytes(), nil
}

func DecompressBody(contentEncoding string, body io.Reader) error {
	var arrBody []byte
	if strings.Contains(contentEncoding, "gzip") {
		bytBody, err := io.ReadAll(body)
		if err != nil {
			return err
		}
		arrBody, err = Decompress(bytBody)
		if err != nil {
			return err
		}

		body = bytes.NewReader(arrBody)
	}
	return nil
}
