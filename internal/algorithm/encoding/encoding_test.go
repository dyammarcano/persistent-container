package encoding

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestEncoding1kChars(t *testing.T) {
	serialized, err := Serialize("hello world")
	assert.Nil(t, err)

	deserialized, err := Deserialize(serialized)
	assert.Nil(t, err)

	assert.Equal(t, deserialized, "hello world")

	fmt.Printf("serialized: %s, deserialized: %s\n", serialized, deserialized)
}
