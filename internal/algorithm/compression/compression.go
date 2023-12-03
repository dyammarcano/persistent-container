package compression

import (
	"bytes"
	"compress/gzip"
	"log"
)

func CompressData(data []byte) ([]byte, error) {
	var buffer bytes.Buffer
	gzipWriter := gzip.NewWriter(&buffer)
	if _, err := gzipWriter.Write(data); err != nil {
		return nil, err
	}

	if err := gzipWriter.Close(); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}

func DecompressData(compressedData []byte) ([]byte, error) {
	var buffer bytes.Buffer
	gzipReader, err := gzip.NewReader(bytes.NewReader(compressedData))
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := gzipReader.Close(); err != nil {
			log.Fatalf("error closing gzip reader: %v", err)
		}
	}()

	if _, err = buffer.ReadFrom(gzipReader); err != nil {
		return nil, err
	}
	return buffer.Bytes(), nil
}
