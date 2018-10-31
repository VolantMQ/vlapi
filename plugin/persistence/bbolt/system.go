package persistenceBbolt

import (
	"github.com/VolantMQ/vlapi/plugin/persistence"
	"github.com/coreos/bbolt"
)

type system struct {
	*dbStatus
}

func (s *system) GetInfo() (*persistence.SystemState, error) {
	state := &persistence.SystemState{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		sys := tx.Bucket(bucketSystem)
		if sys == nil {
			return persistence.ErrNotInitialized
		}

		state.Version = string(sys.Get([]byte("version")))

		return nil
	})

	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *system) SetInfo(state *persistence.SystemState) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		sys := tx.Bucket(bucketSystem)
		if sys == nil {
			return persistence.ErrNotInitialized
		}

		if e := sys.Put([]byte("version"), []byte(state.Version)); e != nil {
			return e
		}

		return nil
	})

	return err
}
