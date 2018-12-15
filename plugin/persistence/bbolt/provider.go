package persistenceBbolt

import (
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"

	"github.com/blang/semver"

	"github.com/VolantMQ/vlapi/plugin"
	"github.com/VolantMQ/vlapi/plugin/persistence"
	"github.com/etcd-io/bbolt"
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

const currentVersion = "0.1.0"

// Config configuration of the BoltDB backend
type Config struct {
	File string `json:"file"`
}

type dbStatus struct {
	db   *bbolt.DB // nolint: structcheck
	done chan struct{}
}

type impl struct {
	*vlplugin.SysParams
	dbStatus

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
func Load(c interface{}, params *vlplugin.SysParams) (persistence.IFace, error) {
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
			params.Log.Infof("creating dir %s", dir)
			if err = os.MkdirAll(dir, os.ModePerm); err != nil {
				return nil, err
			}
		} else if !stat.IsDir() {
			return nil, errors.New("persistence.bbolt: path [" + dir + "] is not directory")
		}
	}

	persist := &impl{
		SysParams: params,
		dbStatus: dbStatus{
			done: make(chan struct{}),
		},
	}

	var err error

	if persist.db, err = bbolt.Open(config.File, 0600, nil); err != nil {
		return nil, err
	}

	err = persist.db.Update(func(tx *bbolt.Tx) error {
		var e error
		if sys := tx.Bucket(bucketSystem); sys == nil {
			if sys, e = tx.CreateBucketIfNotExists(bucketSystem); e != nil {
				return e
			}

			if e = sys.Put([]byte("version"), []byte(currentVersion)); e != nil {
				return e
			}

			if _, e = tx.CreateBucketIfNotExists(bucketSessions); e != nil {
				return e
			}

			if _, e = tx.CreateBucketIfNotExists(bucketRetained); e != nil {
				return e
			}

			params.Log.Info("initializing storage version: ", currentVersion)
		} else {
			data := sys.Get([]byte("version"))

			var ver semver.Version

			if ver, e = semver.Make(string(data)); e != nil {
				e = errors.New("storage is broken. no version field")
				params.Log.Error(e.Error())
				return e
			}

			if sys = tx.Bucket(bucketSessions); sys == nil {
				e = errors.New("storage is broken. no bucket sessions")
				return e
			}

			if sys = tx.Bucket(bucketRetained); sys == nil {
				e = errors.New("storage is broken. no bucket retained")
				return e
			}

			params.Log.Info("using initialized storage: version ", ver.String())
		}

		return e
	})

	if err != nil {
		return nil, err
	}

	persist.r = &retained{
		dbStatus: &persist.dbStatus,
	}

	persist.s = &sessions{
		dbStatus: &persist.dbStatus,
	}

	persist.sys = &system{
		dbStatus: &persist.dbStatus,
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
	select {
	case <-p.done:
		return persistence.ErrNotOpen
	default:
		close(p.done)
	}

	err := p.db.Close()
	p.db = nil

	return err
}

// itob64 returns an 8-byte big endian representation of v
func itob64(v uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	return b
}
