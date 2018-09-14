package main

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"sync"

	"github.com/VolantMQ/vlapi/plugin"
	"github.com/VolantMQ/vlapi/plugin/persistence"
	bolt "github.com/coreos/bbolt"
)

var (
	bucketRetained      = []byte("retained")
	bucketSessions      = []byte("sessions")
	bucketSubscriptions = []byte("subscriptions")
	bucketExpire        = []byte("expire")
	bucketPackets       = []byte("packets")
	bucketPacketsQoS0   = []byte("packetsQoS0")
	bucketPacketsQoS12  = []byte("packetsQoS12")
	bucketPacketsUnAck  = []byte("packetsUnAck")
	bucketState         = []byte("state")
	bucketSystem        = []byte("system")
	sessionsCount       = []byte("count")
)

// Config configuration of the BoltDB backend
type Config struct {
	File string `json:"file"`
}

type dbStatus struct {
	db   *bolt.DB // nolint: structcheck
	done chan struct{}
}

type impl struct {
	*vlplugin.SysParams
	dbStatus

	// transactions that are in progress right now
	wgTx sync.WaitGroup
	lock sync.Mutex

	r   *retained
	s   *sessions
	sys *system
}

var initialBuckets = [][]byte{
	bucketRetained,
	bucketSessions,
	bucketSystem,
}

// nolint: golint
func Load(c interface{}, params *vlplugin.SysParams) (interface{}, error) {
	config, ok := c.(*Config)
	if !ok {
		cfg, kk := c.(map[interface{}]interface{})
		if !kk {
			return nil, persistence.ErrInvalidConfig
		}

		var fileName interface{}
		if fileName, ok = cfg["file"]; !ok {
			return nil, persistence.ErrInvalidConfig
		}

		if _, ok = fileName.(string); !ok {
			return nil, persistence.ErrInvalidConfig
		}

		config = &Config{
			File: fileName.(string),
		}
	}

	if dir := filepath.Dir(config.File); dir != "." {
		if stat, err := os.Stat(dir); os.IsNotExist(err) {
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return nil, err
			}
		} else if !stat.IsDir() {
			return nil, errors.New("persistence.boltdb: path [" + dir + "] is not directory")
		}
	}

	persist := &impl{
		SysParams: params,
		dbStatus: dbStatus{
			done: make(chan struct{}),
		},
	}

	var err error

	if persist.db, err = bolt.Open(config.File, 0600, nil); err != nil {
		return nil, err
	}

	err = persist.db.Update(func(tx *bolt.Tx) error {
		for _, b := range initialBuckets {
			if _, e := tx.CreateBucketIfNotExists(b); e != nil {
				return e
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	persist.r = &retained{
		dbStatus: &persist.dbStatus,
		wgTx:     &persist.wgTx,
		lock:     &persist.lock,
	}

	persist.s = &sessions{
		dbStatus: &persist.dbStatus,
		wgTx:     &persist.wgTx,
		lock:     &persist.lock,
	}

	persist.sys = &system{
		dbStatus: &persist.dbStatus,
		wgTx:     &persist.wgTx,
		lock:     &persist.lock,
	}

	if err = persist.s.init(); err != nil {
		return nil, err
	}

	return persist, nil
}

func (p *impl) System() (persistence.System, error) {
	select {
	case <-p.done:
		return nil, persistence.ErrNotOpen
	default:
	}

	return p.sys, nil
}

// Sessions
func (p *impl) Sessions() (persistence.Sessions, error) {
	select {
	case <-p.done:
		return nil, persistence.ErrNotOpen
	default:
	}

	return p.s, nil
}

// Retained
func (p *impl) Retained() (persistence.Retained, error) {
	select {
	case <-p.done:
		return nil, persistence.ErrNotOpen
	default:
	}

	return p.r, nil
}

// Shutdown provider
func (p *impl) Shutdown() error {
	p.lock.Lock()
	defer p.lock.Unlock()

	select {
	case <-p.done:
		return persistence.ErrNotOpen
	default:
	}

	close(p.done)

	p.wgTx.Wait()

	err := p.db.Close()
	p.db = nil

	return err
}

// itob64 returns an 8-byte big endian representation of v.
func itob64(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
