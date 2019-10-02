package persistenceBbolt

import (
	"github.com/coreos/bbolt"

	"github.com/VolantMQ/vlapi/vlplugin/vlpersistence"
)

type system struct {
	*dbStatus
}

func (s *system) GetInfo() (*vlpersistence.SystemState, error) {
	state := &vlpersistence.SystemState{}

	err := s.db.View(func(tx *bbolt.Tx) error {
		sys := tx.Bucket(bucketSystem)
		if sys == nil {
			return vlpersistence.ErrNotInitialized
		}

		state.Version = string(sys.Get([]byte("version")))

		return nil
	})

	if err != nil {
		return nil, err
	}

	return state, nil
}

func (s *system) SetInfo(state *vlpersistence.SystemState) error {
	err := s.db.Update(func(tx *bbolt.Tx) error {
		sys := tx.Bucket(bucketSystem)
		if sys == nil {
			return vlpersistence.ErrNotInitialized
		}

		if e := sys.Put([]byte("version"), []byte(state.Version)); e != nil {
			return e
		}

		return nil
	})

	return err
}
