package transform

import (
	"context"
	"dataStore/internal/persistence"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

type testStruct struct {
	Name     string
	LastName string
	Age      int
}

func TestNewContainer(t *testing.T) {
	db, err := persistence.NewPersistence(context.TODO(), "testNewContainer.db")
	assert.NoErrorf(t, err, "error creating persistence")
	assert.NotNil(t, db, "persistence is nil")

	container := NewContainer[testStruct](testStruct{
		Name:     "John",
		LastName: "Wick",
		Age:      42,
	}, "test", "testKey", db)
	assert.NotNil(t, container, "container is nil")

	objectInstance := container.GetObject()

	objectInstance.Age = 43
	objectInstance.Name = "John2"
	objectInstance.LastName = "Wick2"

	<-time.After(5 * time.Second)

	assert.Equalf(t, 43, container.GetObject().Age, "Age not equal")
}
