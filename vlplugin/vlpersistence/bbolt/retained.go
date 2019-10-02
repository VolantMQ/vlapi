package persistenceBbolt

import (
	"github.com/coreos/bbolt"

	"github.com/VolantMQ/vlapi/vlplugin/vlpersistence"
)

type retained struct {
	*dbStatus
}

func (r *retained) Load() ([]*vlpersistence.PersistedPacket, error) {
	var res []*vlpersistence.PersistedPacket

	err := r.db.View(func(tx *bbolt.Tx) error {
		bucket := tx.Bucket(bucketRetained)

		return bucket.ForEach(func(k, v []byte) error {
			pkt := &vlpersistence.PersistedPacket{}

			if buck := bucket.Bucket(k); buck != nil {
				pkt.Data = buck.Get([]byte("data"))
				pkt.ExpireAt = string(buck.Get([]byte("expireAt")))

				res = append(res, pkt)
			}

			return nil
		})
	})

	return res, err
}

// Store retained packets
func (r *retained) Store(packets []*vlpersistence.PersistedPacket) error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket(bucketRetained); err != nil && err != bbolt.ErrBucketNotFound {
			return err
		}

		bucket, err := tx.CreateBucket(bucketRetained)
		if err != nil {
			return err
		}

		for _, p := range packets {
			var id uint64
			if id, err = bucket.NextSequence(); err != nil {
				return err
			}
			var pack *bbolt.Bucket
			if pack, err = bucket.CreateBucketIfNotExists(itob64(id)); err != nil {
				return err
			}

			if err = pack.Put([]byte("data"), p.Data); err != nil {
				return err
			}

			if err = pack.Put([]byte("expireAt"), []byte(p.ExpireAt)); err != nil {
				return err
			}
		}

		return nil
	})
}

// Wipe all retained packets
func (r *retained) Wipe() error {
	return r.db.Update(func(tx *bbolt.Tx) error {
		if err := tx.DeleteBucket(bucketRetained); err != nil {
			return err
		}

		if _, err := tx.CreateBucket(bucketRetained); err != nil {
			return err
		}
		return nil
	})
}
