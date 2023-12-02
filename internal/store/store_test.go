package store

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewPersistence(t *testing.T) {
	per, err := NewPersistence("test.db")
	assert.NoErrorf(t, err, "error creating store")
	assert.NotNil(t, per, "store is nil")

	defer per.Close()

	// setting key/value pair
	key := "Rogers"
	value := []byte("Avengers, Assemble!")
	movie := "Captain America"

	err = per.Put(movie, key, value)
	assert.NoErrorf(t, err, "error putting key/value pair")

	// getting value from key
	data, err := per.Get(movie, key)
	assert.NoErrorf(t, err, "error getting value from key")

	fmt.Printf("data: %s\n", data)

	// batch insert
	values := make([][]byte, 100)

	for i := 0; i < 100; i++ {
		values = append(values, []byte(fmt.Sprintf("value %d", i)))
	}

	err = per.PutBatch(movie, key, values)
	assert.NoErrorf(t, err, "error putting batch key/value pair")

	// getting value from key
	obj, err := per.GetBucketValues(movie)
	assert.NoErrorf(t, err, "error getting value from key")

	for _, v := range obj {
		fmt.Printf("data: %s\n", v)
	}
}

//func testDBAction(t *testing.T, action func(*Store) error) {
//	tmpDir, _ := os.MkdirTemp("", "prefix")
//	defer os.Remove(tmpDir) // clean up
//
//	p, _ := NewPersistence(tmpDir)
//	defer p.Close() // clean up
//
//	assert.Nil(t, action(p))
//
//	p.DB.View(func(tx *bbolt.Tx) error {
//		b := tx.Bucket([]byte("testBucket"))
//		assert.NotNil(t, b)
//		return nil
//	})
//}
//
//func TestClose(t *testing.T) {
//	tmpDir, _ := os.MkdirTemp("", "prefix")
//	defer os.Remove(tmpDir) // clean up
//
//	p, _ := NewPersistence(tmpDir)
//	assert.Nil(t, p.Close())
//}

//func TestUpdate(t *testing.T) {
//	testDBAction(t, func(p *Store) error {
//		return p.Update(func(tx *bbolt.Tx) error {
//			_, err := tx.CreateBucket([]byte("testBucket"))
//			return err
//		})
//	})
//}
//
//func TestView(t *testing.T) {
//	testDBAction(t, func(p *Store) error {
//		return p.View(func(tx *bbolt.Tx) error {
//			_ = tx.Bucket([]byte("testBucket"))
//			return nil
//		})
//	})
//}
//
//func TestBatch(t *testing.T) {
//	testDBAction(t, func(p *Store) error {
//		return p.Batch(func(tx *bbolt.Tx) error {
//			_, err := tx.CreateBucket([]byte("testBucket"))
//			return err
//		})
//	})
//}
