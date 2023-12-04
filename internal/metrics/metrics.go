package metrics

import (
	"context"
	"encoding/json"
	"errors"
	bolt "go.etcd.io/bbolt"
	"sync"
	"time"
)

type (
	Metrics struct {
		Uptime        time.Time     `json:"uptime"`
		LastUpdate    time.Time     `json:"last_update"`
		SystemMetrics SystemMetrics `json:"system_metrics"`
		Iops          Iops          `json:"iops"`
		db            *bolt.DB
		mu            sync.RWMutex
		ctx           context.Context
		bucketName    []byte
		readsCh       chan int64
		writesCh      chan int64
	}

	SystemMetrics struct {
		TotalBuckets int64         `json:"total_buckets"`
		TotalKeys    int64         `json:"total_keys"`
		Hits         int64         `json:"hits"`
		Miss         int64         `json:"miss"`
		Pages        int64         `json:"pages"`
		Stats        *bolt.TxStats `json:"stats"`
	}

	Iops struct {
		TotalReads      int64 `json:"total_reads"`
		TotalWrites     int64 `json:"total_writes"`
		WritesPerSecond int64 `json:"writes_per_second"`
		ReadsPerSecond  int64 `json:"reads_per_second"`
		TotalReadBytes  int64 `json:"total_read_bytes"`
		TotalWriteBytes int64 `json:"total_write_bytes"`
	}
)

func NewMetrics(ctx context.Context, db *bolt.DB) *Metrics {
	m := &Metrics{
		ctx:        ctx,
		db:         db,
		bucketName: []byte("metrics"),
		readsCh:    make(chan int64, 100),
		writesCh:   make(chan int64, 100),
		Uptime:     time.Now(),
		LastUpdate: time.Now(),
		SystemMetrics: SystemMetrics{
			TotalBuckets: 0,
			TotalKeys:    0,
			Hits:         0,
			Miss:         0,
			Pages:        0,
			Stats:        &bolt.TxStats{},
		},
		Iops: Iops{
			TotalReads:      0,
			TotalWrites:     0,
			WritesPerSecond: 0,
			ReadsPerSecond:  0,
			TotalReadBytes:  0,
			TotalWriteBytes: 0,
		},
	}

	go m.startMonitor()

	return m
}

func (m *Metrics) startMonitor() {
	ticker := time.NewTicker(time.Second * 1)
	defer ticker.Stop()

	readSavedData := func() {
		err := m.db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket(m.bucketName)
			if bucket == nil {
				return errors.New("metrics bucket not found")
			}

			data := bucket.Get(m.bucketName)
			if data == nil {
				return errors.New("metrics key not found")
			}

			m.mu.Lock()
			err := json.Unmarshal(data, m)
			m.mu.Unlock()

			if err != nil {
				return errors.New("failed to unmarshal metrics data: " + err.Error())
			}

			return nil
		})
		if err != nil {
			return
		}
	}

	readSavedData()

	for {
		select {
		case <-m.ctx.Done():
			return
		case <-ticker.C:
			m.resetIopsRates()
		case v := <-m.readsCh:
			m.incrementTotalReadBytes(v)
		case v := <-m.writesCh:
			m.incrementTotalWriteBytes(v)
		}
	}
}

func (m *Metrics) resetIopsRates() {
	m.mu.Lock()
	defer m.mu.Unlock()

	saveData, err := json.Marshal(m)
	if err != nil {
		return
	}

	err = m.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists(m.bucketName)
		if err != nil {
			return err
		}

		return bucket.Put(m.bucketName, saveData)
	})
	if err != nil {
		return
	}

	m.Iops.ReadsPerSecond = 0
	m.Iops.WritesPerSecond = 0
}

func (m *Metrics) incrementTotalReadBytes(v int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Iops.TotalReadBytes += v
}

func (m *Metrics) incrementTotalWriteBytes(v int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Iops.TotalWriteBytes += v
}

func (m *Metrics) UpdateMetrics(tx *bolt.Tx) {
	m.mu.Lock()
	defer m.mu.Unlock()

	stats := tx.Stats()
	m.SystemMetrics.Stats = &stats
}

func (m *Metrics) UpdateIopsReads() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Iops.TotalReads++
	m.Iops.ReadsPerSecond++
}

func (m *Metrics) UpdateIopsWrites() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.Iops.TotalWrites++
	m.Iops.WritesPerSecond++
}

func (m *Metrics) GetMetrics() (*Metrics, error) {
	err := m.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket(m.bucketName)
		if bucket == nil {
			return errors.New("metrics bucket not found")
		}

		data := bucket.Get(m.bucketName)
		if data == nil {
			return errors.New("metrics key not found")
		}

		m.mu.Lock()
		err := json.Unmarshal(data, m)
		m.mu.Unlock()

		if err != nil {
			return errors.New("failed to unmarshal metrics data: " + err.Error())
		}

		m.mu.Lock()
		m.LastUpdate = time.Now()
		m.Iops.TotalReads++
		m.mu.Unlock()

		return nil
	})

	return m, err
}
