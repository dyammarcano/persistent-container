package persistence

import (
	"encoding/json"
	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
	"sync"
)

type (
	performAction func(tx *bolt.Tx) error

	Persistence struct {
		*bolt.DB
		mu sync.RWMutex
	}
)

func NewPersistence(path string) (*Persistence, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	return &Persistence{
		DB: db,
		mu: sync.RWMutex{},
	}, nil
}

func (p *Persistence) Close() error {
	return p.DB.Close()
}

func (p *Persistence) Update(fn performAction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.DB.Update(fn)
}

func (p *Persistence) View(fn performAction) error {
	return p.DB.View(fn)
}

func (p *Persistence) Batch(fn performAction) error {
	return p.DB.Batch(fn)
}

func (p *Persistence) DeleteBucket(bucketName string) error {
	return p.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(bucketName))
	})
}

func (p *Persistence) DeleteKey(bucketName string, key string) error {
	return p.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}
		return bucket.Delete([]byte(key))
	})
}

func (p *Persistence) Put(bucketName string, key string, value []byte) error {
	return p.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		return bucket.Put([]byte(key), value)
	})
}

func (p *Persistence) PutBatch(bucketName string, key string, values [][]byte) error {
	return p.Batch(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}
		for _, value := range values {
			if err = bucket.Put([]byte(key), value); err != nil {
				return err
			}
		}
		return nil
	})
}

func (p *Persistence) Get(bucketName string, key string) ([]byte, error) {
	var value []byte
	err := p.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}
		value = bucket.Get([]byte(key))
		return nil
	})
	return value, err
}

func (p *Persistence) GetBucket(bucketName string) (*bolt.Bucket, error) {
	var bucket *bolt.Bucket
	err := p.View(func(tx *bolt.Tx) error {
		bucket = tx.Bucket([]byte(bucketName))
		return nil
	})
	return bucket, err
}

func (p *Persistence) GetBucketKeys(bucketName string) ([]string, error) {
	var keys []string
	err := p.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		})
	})
	return keys, err
}

func (p *Persistence) GetBucketValues(bucketName string) ([][]byte, error) {
	var values [][]byte
	err := p.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(k, v []byte) error {
			values = append(values, v)
			return nil
		})
	})
	return values, err
}

func (p *Persistence) GetBucketKeysValues(bucketName string) ([][]byte, [][]byte, error) {
	var keys [][]byte
	var values [][]byte
	err := p.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}
		return bucket.ForEach(func(k, v []byte) error {
			keys = append(keys, k)
			values = append(values, v)
			return nil
		})
	})
	return keys, values, err
}

func (p *Persistence) SerializeObject(obj any) ([]byte, error) {
	return json.Marshal(obj)
}

func (p *Persistence) DeserializeObject(data []byte, obj any) error {
	return json.Unmarshal(data, obj)
}

func (p *Persistence) GenerateKey() string {
	return uuid.New().String()
}

func (p *Persistence) GenerateKeyBytes() []byte {
	return []byte(p.GenerateKey())
}
