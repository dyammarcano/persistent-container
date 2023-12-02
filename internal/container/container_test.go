package container

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

type testStruct struct {
	Name     string
	LastName string
	Legend   bool
	Age      int
}

func TestNewContainer(t *testing.T) {
	expectedObj := testStruct{
		Name:     "Chuck",
		LastName: "Norris",
		Legend:   true,
		Age:      42,
	}

	expectedName := "Test"
	container1 := NewContainer(expectedName, expectedObj)

	// Test if container is not nil
	assert.NotNilf(t, container1, "NewContainer() = nil; want non-nil")

	// Test if UID is not empty
	assert.NotEmptyf(t, container1.GetUid(), "GetUid() = ''; want non-empty string")

	// Test if container.Name is equal to expected name
	assert.Equalf(t, container1.Name, expectedName, "Name not equal")

	// Test if the object in container is equal to expected object
	assert.Equalf(t, container1.GetObject(), expectedObj, "Object not equal")

	// Test if the timestamp is not 0
	assert.NotZero(t, container1.GetTimestamp(), "Timestamp is 0; want non-zero")

	data, err := container1.Pack()
	assert.NoErrorf(t, err, "Error encoding data: %v", err)

	container2 := NewContainer(expectedName, testStruct{})

	err = container2.UnPack(data)
	assert.NoErrorf(t, err, "Error decoding data: %v", err)

	assert.Equalf(t, container2.Name, container1.Name, "Name not equal")
}
