package container

import (
	"context"
	"dataStore/internal/store"
	"github.com/fxamacker/cbor/v2"
	"github.com/google/uuid"
	"reflect"
	"sync"
	"time"
)

type (
	Container[T any] struct {
		uid         string
		timestamp   int64
		item        T
		itemClone   T
		bucketName  string
		key         string
		db          *store.Store
		saved       bool
		modified    bool
		lasModified int64
		ctx         context.Context
		mu          sync.RWMutex
	}
)

func NewContainer[T any](obj T, bucketName string, key string, db *store.Store) *Container[T] {
	c := &Container[T]{
		db:         db,
		item:       obj,
		ctx:        db.Ctx,
		bucketName: bucketName,
		key:        key,
		mu:         sync.RWMutex{},
		uid:        uuid.NewString(),
		timestamp:  time.Now().UnixNano(),
	}

	c.clone()         // clone object
	go c.isModified() // start isModified goroutine

	return c
}

// encode encodes the object to cbor
func (c *Container[T]) encode() ([]byte, error) {
	return cbor.Marshal(c.item)
}

// decode decodes the object from cbor
func (c *Container[T]) decode(data []byte) error {
	if err := cbor.Unmarshal(data, &c.item); err != nil {
		return err
	}
	c.clone()
	return nil
}

// clone clones the object
func (c *Container[T]) clone() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.itemClone = c.item
}

// isModified checks if the object has been modified
func (c *Container[T]) isModified() {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case tick := <-ticker.C:
			c.checkModified(tick)
		case <-c.ctx.Done():
			return
		}
	}
}

// checkModified checks if the object has been modified
func (c *Container[T]) checkModified(tick time.Time) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if tick.Unix()-c.lasModified > 1 {
		c.modified = !reflect.DeepEqual(c.item, c.itemClone)
		c.lasModified = tick.Unix()

		if c.modified {
			if err := c.set(); err != nil {
				return
			}
		}
	}
}

// get returns the object from the database
func (c *Container[T]) get() error {
	data, err := c.db.Get(c.bucketName, c.key)
	if err != nil {
		return err
	}

	return c.decode(data)
}

// set sets the object in the database
func (c *Container[T]) set() error {
	if !c.modified {
		return nil // No need to set if not modified
	}

	data, err := c.encode()
	if err != nil {
		return err
	}

	if err = c.db.Put(c.bucketName, c.key, data); err != nil {
		return err
	}

	c.saved = true
	c.modified = false
	c.clone()

	return nil
}

// Save saves the object to the database
func (c *Container[T]) Save() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Check if the key exists in the database
	if err := c.get(); err != nil {
		return err
	}

	// Update the database only if modified
	return c.set()
}

// IsModified returns true if the object has been modified
func (c *Container[T]) IsModified() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.modified
}

// IsSaved returns true if the object has been saved
func (c *Container[T]) IsSaved() bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.saved
}

// GetObject returns the object
func (c *Container[T]) GetObject() *T {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return &c.item
}

// GetUid returns the uid
func (c *Container[T]) GetUid() string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.uid
}

// GetTimestamp returns the timestamp
func (c *Container[T]) GetTimestamp() int64 {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.timestamp
}
