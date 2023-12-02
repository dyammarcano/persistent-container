package container

import (
	"dataStore/internal/transform"
	"github.com/google/uuid"
	"time"
)

type Container[T any] struct {
	UID       string
	Timestamp int64
	Name      string
	Object    T
}

func NewContainer[T any](name string, obj T) *Container[T] {
	return &Container[T]{
		Timestamp: time.Now().UnixNano(),
		UID:       uuid.NewString(),
		Name:      name,
		Object:    obj,
	}
}

func (c *Container[T]) GetUid() string {
	return c.UID
}

func (c *Container[T]) GetObject() T {
	return c.Object
}

func (c *Container[T]) GetName() string {
	return c.Name
}

func (c *Container[T]) GetTimestamp() int64 {
	return c.Timestamp
}

func (c *Container[T]) Pack() ([]byte, error) {
	return transform.EncodeGob(c)
}

func (c *Container[T]) UnPack(data []byte) error {
	return transform.DecodeGob(data, c)
}
