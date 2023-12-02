package transform

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testStruct struct {
	Name     string
	LastName string
	Age      int
}

func TestDecodeGob(t *testing.T) {
	testData := testStruct{
		Name:     "John",
		LastName: "Wick",
		Age:      42,
	}

	encoded, err := EncodeGob(testData)
	assert.NoErrorf(t, err, "Error encoding data: %v", err)

	//var decoded testStruct
	decoded, err := DecodeGobGeneric[testStruct](encoded)
	assert.NoErrorf(t, err, "Error decoding data: %v", err)

	assert.Equalf(t, testData.Age, decoded.Age, "Age not equal")
	assert.Equalf(t, testData.Name, decoded.Name, "Name not equal")
}
