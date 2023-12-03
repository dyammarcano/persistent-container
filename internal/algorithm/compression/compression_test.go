package compression

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCompressionAndDecompression(t *testing.T) {
	data := []byte("Hello, World!")

	// Test compression
	compressedData, err := CompressData(data)
	assert.Nil(t, err, "CompressData failed")

	// Test decompression
	decompressedData, err := DecompressData(compressedData)
	assert.Nil(t, err, "DecompressData failed")

	assert.Equal(t, data, decompressedData, "Decompressed data is not equal to original data")
}
