package store

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dyammarcano/persistent-container/internal/algorithm/compression"
	"github.com/dyammarcano/persistent-container/internal/metrics"
	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"
	"sync"
	"time"
)

type (
	performAction func(tx *bolt.Tx) error

	Store struct {
		*bolt.DB
		mu      sync.RWMutex
		Ctx     context.Context
		metrics *metrics.Metrics
	}

	Key struct {
		Key string `json:"id"`
	}

	WrapData struct {
		Object    any
		Timestamp int64
		Data      []byte
	}
)

func NewStore(ctx context.Context, path string) (*Store, error) {
	db, err := bolt.Open(path, 0600, nil)
	if err != nil {
		return nil, err
	}

	s := &Store{
		DB:  db,
		Ctx: ctx,
		mu:  sync.RWMutex{},
	}

	s.metrics = metrics.NewMetrics(ctx, db)

	return s, nil
}

func (p *Store) GetMetrics() *metrics.Metrics {
	p.mu.RLock()
	defer p.mu.RUnlock()

	getMetrics, err := p.metrics.GetMetrics()
	if err != nil {
		return nil
	}
	return getMetrics
}

func (p *Store) Close() error {
	return p.DB.Close()
}

func (p *Store) Update(fn performAction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.metrics.LastUpdate = time.Now()
	p.metrics.Iops.TotalWrites++

	return p.DB.Update(fn)
}

func (p *Store) View(fn performAction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.metrics.LastUpdate = time.Now()
	p.metrics.Iops.TotalReads++

	return p.DB.View(fn)
}

func (p *Store) Batch(fn performAction) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	return p.DB.Batch(fn)
}

func (p *Store) DeleteBucket(bucketName string) error {
	return p.Update(func(tx *bolt.Tx) error {
		return tx.DeleteBucket([]byte(bucketName))
	})
}

func (p *Store) DeleteKey(bucketName string, key string) error {
	return p.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			return nil
		}
		return bucket.Delete([]byte(key))
	})
}

func (p *Store) Put(bucketName string, key string, value []byte) error {
	return p.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucketName))
		if err != nil {
			return err
		}

		p.metrics.Iops.TotalWriteBytes += int64(len(value))

		return bucket.Put([]byte(key), value)
	})
}

func (p *Store) PutBatch(bucketName string, key string, values [][]byte) error {
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

func (p *Store) Get(bucketName string, key string) ([]byte, error) {
	var value []byte
	err := p.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			value = []byte{}
			return nil
		}
		value = bucket.Get([]byte(key))
		return nil
	})
	return value, err
}

func (p *Store) GetBucket(bucketName string) (*bolt.Bucket, error) {
	var bucket *bolt.Bucket
	err := p.View(func(tx *bolt.Tx) error {
		bucket = tx.Bucket([]byte(bucketName))
		return nil
	})
	return bucket, err
}

func (p *Store) GetBucketKeys(bucketName string) ([]Key, error) {
	var keys []Key
	err := p.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucketName))
		if bucket == nil {
			keys = []Key{}
			return nil
		}
		return bucket.ForEach(func(k, v []byte) error {
			keys = append(keys, Key{Key: string(k)})
			return nil
		})
	})
	return keys, err
}

func (p *Store) GetBucketValues(bucketName string) ([][]byte, error) {
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

func (p *Store) GetBucketKeysValues(bucketName string) ([][]byte, [][]byte, error) {
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

func (p *Store) PutObject(bucketName string, key string, w *WrapData) error {
	data, err := json.Marshal(w)
	if err != nil {
		return err
	}

	comp, err := compression.CompressData(data)
	if err != nil {
		return err
	}

	return p.Put(bucketName, key, comp)
}

func (p *Store) GetObject(bucketName string, key string) (*WrapData, error) {
	value, err := p.Get(bucketName, key)
	if err != nil {
		return nil, err
	}

	dec, err := compression.DecompressData(value)
	if err != nil {
		return nil, err
	}

	var w WrapData
	if err = json.Unmarshal(dec, &w); err != nil {
		return nil, fmt.Errorf("failed to decode Object: %v", err)
	}
	return &w, err
}

func GenerateKey() string {
	return uuid.NewString()
}
