package main

import (
	"encoding/binary"
	"errors"
	"sync"

	"github.com/VolantMQ/vlapi/plugin/persistence"
	bolt "github.com/coreos/bbolt"
)

type sessions struct {
	*dbStatus

	// transactions that are in progress right now
	wgTx *sync.WaitGroup
	lock *sync.Mutex
}

var _ persistence.Sessions = (*sessions)(nil)

func (s *sessions) init() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		val := sessions.Get(sessionsCount)
		if len(val) == 0 {
			buf := [8]byte{}
			num := binary.PutUvarint(buf[:], 0)
			sessions.Put(sessionsCount, buf[:num]) // nolint: errcheck
		}

		return nil
	})
}

func (s *sessions) Exists(id []byte) bool {
	err := s.db.View(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		ses := sessions.Bucket(id)
		if ses == nil {
			return bolt.ErrBucketNotFound
		}

		return nil
	})

	return err == nil
}

func (s *sessions) Count() uint64 {
	var count uint64
	s.db.View(func(tx *bolt.Tx) error { // nolint: errcheck
		sessions := tx.Bucket(bucketSessions)

		val := sessions.Get(sessionsCount)
		if cnt, num := binary.Uvarint(val); num > 0 {
			count = cnt
		}

		return nil
	})

	return count
}

func (s *sessions) SubscriptionsStore(id []byte, data []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		session, err := getSession(id, sessions)
		if err != nil {
			return err
		}

		return session.Put(bucketSubscriptions, data)
	})
}

func (s *sessions) SubscriptionsDelete(id []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		session := tx.Bucket(id)
		if session == nil {
			return persistence.ErrNotFound
		}

		session.Delete(bucketSubscriptions) // nolint: errcheck
		return nil
	})
}

func boolToByte(v bool) byte {
	if v {
		return 1
	}
	return 0
}

func byteToBool(v byte) bool {
	return !(v == 0)
}

func (s *sessions) PacketsForEachQoS0(id []byte, ctx interface{}, load persistence.PacketLoader) error {
	return s.packetsForEach(id, ctx, load, bucketPacketsQoS0)
}

func (s *sessions) PacketsForEachQoS12(id []byte, ctx interface{}, load persistence.PacketLoader) error {
	return s.packetsForEach(id, ctx, load, bucketPacketsQoS12)
}

func (s *sessions) PacketsForEachUnAck(id []byte, ctx interface{}, load persistence.PacketLoader) error {
	return s.packetsForEach(id, ctx, load, bucketPacketsUnAck)
}

func (s *sessions) packetsForEach(id []byte, ctx interface{}, load persistence.PacketLoader, bucket []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		root := tx.Bucket(bucketSessions)
		if root == nil {
			return persistence.ErrNotInitialized
		}

		session := root.Bucket(id)
		if session == nil {
			return nil
		}

		packetsRoot := session.Bucket(bucketPackets)
		if packetsRoot == nil {
			return nil
		}

		packets := packetsRoot.Bucket(bucket)
		if packets == nil {
			return nil
		}

		packets.ForEach(func(k, v []byte) error { // nolint: errcheck
			if packet := packets.Bucket(k); packet != nil {
				pPkt := &persistence.PersistedPacket{}

				if data := packet.Get([]byte("data")); len(data) > 0 {
					pPkt.Data = data
				} else {
					return persistence.ErrBrokenEntry
				}

				if data := packet.Get([]byte("expireAt")); len(data) > 0 {
					pPkt.ExpireAt = string(data)
				}

				rm, err := load(ctx, pPkt)
				if rm {
					packets.DeleteBucket(k) // nolint: errcheck
				}

				return err
			}
			return nil
		})

		return nil
	})
}

func (s *sessions) PacketCountQoS0(id []byte) (int, error) {
	return s.packetCount(id, bucketPacketsQoS0)
}

func (s *sessions) PacketCountQoS12(id []byte) (int, error) {
	return s.packetCount(id, bucketPacketsQoS12)
}

func (s *sessions) PacketCountUnAck(id []byte) (int, error) {
	return s.packetCount(id, bucketPacketsUnAck)
}

func (s *sessions) packetCount(id []byte, bucket []byte) (int, error) {
	var count int

	err := s.db.View(func(tx *bolt.Tx) error {
		root := tx.Bucket(bucketSessions)
		if root == nil {
			return persistence.ErrNotInitialized
		}

		session := root.Bucket(id)
		if session == nil {
			return nil
		}

		packetsRoot := session.Bucket(bucketPackets)
		if packetsRoot == nil {
			return nil
		}

		packets := packetsRoot.Bucket(bucket)
		if packets == nil {
			return nil
		}

		count = packets.Stats().InlineBucketN

		return nil
	})

	return count, err
}

func (s *sessions) PacketsStore(id []byte, packets persistence.PersistedPackets) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if len(packets.UnAck) > 0 {
			if err := s.packetsStore(tx, id, bucketPacketsUnAck, packets.UnAck); err != nil {
				return err
			}
		}

		if len(packets.QoS0) > 0 {
			if err := s.packetsStore(tx, id, bucketPacketsQoS0, packets.QoS0); err != nil {
				return err
			}
		}

		if len(packets.QoS12) > 0 {
			if err := s.packetsStore(tx, id, bucketPacketsQoS12, packets.QoS12); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *sessions) PacketStoreQoS0(id []byte, pkt *persistence.PersistedPacket) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return s.packetsStore(tx, id, bucketPacketsQoS0, []*persistence.PersistedPacket{pkt})
	})
}

func (s *sessions) PacketStoreQoS12(id []byte, pkt *persistence.PersistedPacket) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return s.packetsStore(tx, id, bucketPacketsQoS12, []*persistence.PersistedPacket{pkt})
	})
}

func (s *sessions) PacketStoreUnAck(id []byte, pkt *persistence.PersistedPacket) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		return s.packetsStore(tx, id, bucketPacketsUnAck, []*persistence.PersistedPacket{pkt})
	})
}

func (s *sessions) packetsStore(tx *bolt.Tx, id []byte, bucket []byte, packets []*persistence.PersistedPacket) error {
	buck, err := createPacketsBucket(tx, id, bucket)
	if err != nil {
		return err
	}

	for _, entry := range packets {
		if err = storePacket(buck, entry); err != nil {
			return err
		}
	}

	return nil
}

func (s *sessions) PacketsDelete(id []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		if sessions := tx.Bucket(bucketSessions); sessions != nil {
			ses := sessions.Bucket(id)
			if ses == nil {
				return persistence.ErrNotFound
			}
			ses.DeleteBucket(bucketPackets) // nolint: errcheck
		}
		return nil
	})
}

func (s *sessions) LoadForEach(loader persistence.SessionLoader, context interface{}) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return nil
		}

		return sessions.ForEach(func(k, v []byte) error {
			// If there's a value, it's not a bucket so ignore it.
			if v != nil {
				return nil
			}

			session := sessions.Bucket(k)
			st := &persistence.SessionState{}

			state := session.Bucket(bucketState)
			if state != nil {
				if v := state.Get([]byte("version")); len(v) > 0 {
					st.Version = v[0]
				} else {
					st.Errors = append(st.Errors, errors.New("protocol version not found"))
				}

				st.Timestamp = string(state.Get([]byte("timestamp")))
				st.Subscriptions = state.Get(bucketSubscriptions)

				if expire := state.Bucket(bucketExpire); expire != nil {
					st.Expire = &persistence.SessionDelays{
						Since:    string(expire.Get([]byte("since"))),
						ExpireIn: string(expire.Get([]byte("expireIn"))),
						Will:     expire.Get([]byte("will")),
					}
				}
			}
			return loader.LoadSession(context, k, st)
		})
	})
}

func (s *sessions) Create(id []byte, state *persistence.SessionBase) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		session, err := getSession(id, sessions)
		if err != nil {
			return err
		}

		var st *bolt.Bucket
		st, err = session.CreateBucketIfNotExists(bucketState)
		if err != nil {
			return err
		}

		if err = st.Put([]byte("timestamp"), []byte(state.Timestamp)); err != nil {
			return err
		}

		if err = st.Put([]byte("version"), []byte{state.Version}); err != nil {
			return err
		}

		return nil
	})
}

func (s *sessions) StateStore(id []byte, state *persistence.SessionState) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		session, err := getSession(id, sessions)
		if err != nil {
			return err
		}

		var st *bolt.Bucket
		st, err = session.CreateBucketIfNotExists(bucketState)
		if err != nil {
			return err
		}

		if len(state.Subscriptions) > 0 {
			if err = st.Put(bucketSubscriptions, state.Subscriptions); err != nil {
				return err
			}
		}

		if err = st.Put([]byte("timestamp"), []byte(state.Timestamp)); err != nil {
			return err
		}

		if state.Expire != nil {
			expire, err := st.CreateBucketIfNotExists(bucketExpire)
			if err != nil {
				return err
			}

			if err = expire.Put([]byte("since"), []byte(state.Expire.Since)); err != nil {
				return err
			}

			if len(state.Expire.ExpireIn) > 0 {
				if err = expire.Put([]byte("expireIn"), []byte(state.Expire.ExpireIn)); err != nil {
					return err
				}
			}

			if len(state.Expire.Will) > 0 {
				if err = expire.Put([]byte("will"), state.Expire.Will); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

func (s *sessions) ExpiryStore(id []byte, exp *persistence.SessionDelays) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		session, err := getSession(id, sessions)
		if err != nil {
			return err
		}

		var st *bolt.Bucket
		st, err = session.CreateBucketIfNotExists(bucketState)
		if err != nil {
			return err
		}

		expire, err := st.CreateBucketIfNotExists(bucketExpire)
		if err != nil {
			return err
		}

		if err = expire.Put([]byte("since"), []byte(exp.Since)); err != nil {
			return err
		}

		if len(exp.ExpireIn) > 0 {
			if err = expire.Put([]byte("expireIn"), []byte(exp.ExpireIn)); err != nil {
				return err
			}
		}

		if len(exp.Will) > 0 {
			if err = expire.Put([]byte("will"), exp.Will); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *sessions) ExpiryDelete(id []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return nil
		}

		session := sessions.Bucket(id)
		if session == nil {
			return nil
		}
		state := session.Bucket(bucketState)
		if state == nil {
			return nil
		}

		state.DeleteBucket(bucketExpire) // nolint: errcheck

		return nil
	})
}

func (s *sessions) StateDelete(id []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		session, err := getSession(id, sessions)
		if err != nil {
			return err
		}

		session.DeleteBucket(bucketState) // nolint: errcheck

		return nil
	})
}

func (s *sessions) Delete(id []byte) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		session := sessions.Bucket(id)
		if session == nil {
			return persistence.ErrNotFound
		}

		if err := sessions.DeleteBucket(id); err != nil {
			return err
		}

		count, num := binary.Uvarint(sessions.Get(sessionsCount))
		if num <= 0 {
			return errors.New("persistence: broken count")
		}

		if count == 0 {
			return errors.New("persistence: count already 0")
		}

		count--
		buf := [8]byte{}
		num = binary.PutUvarint(buf[:], count)
		sessions.Put(sessionsCount, buf[:num]) // nolint: errcheck

		return nil
	})
}

func getSession(id []byte, sessions *bolt.Bucket) (*bolt.Bucket, error) {
	session := sessions.Bucket(id)
	if session == nil {
		var err error
		if session, err = sessions.CreateBucket(id); err != nil {
			return nil, err
		}

		count, num := binary.Uvarint(sessions.Get(sessionsCount))
		if num <= 0 {
			sessions.DeleteBucket(id) // nolint: errcheck
			return nil, nil
		}

		count++

		buf := [8]byte{}
		num = binary.PutUvarint(buf[:], count)
		sessions.Put([]byte("count"), buf[:num]) // nolint: errcheck
	}

	return session, nil
}

func createPacketsBucket(tx *bolt.Tx, id []byte, bucket []byte) (*bolt.Bucket, error) {
	sessions, err := tx.CreateBucketIfNotExists(bucketSessions)
	if err != nil {
		return nil, err
	}

	var session *bolt.Bucket
	if session, err = getSession(id, sessions); err != nil {
		return nil, err
	}

	var packetsRoot *bolt.Bucket
	if packetsRoot, err = session.CreateBucketIfNotExists(bucketPackets); err != nil {
		return nil, err
	}

	return packetsRoot.CreateBucketIfNotExists(bucket)
}

func storePacket(buck *bolt.Bucket, packet *persistence.PersistedPacket) error {
	id, _ := buck.NextSequence() // nolint: gas
	pBuck, err := buck.CreateBucketIfNotExists(itob64(id))
	if err != nil {
		return err
	}

	if err = pBuck.Put([]byte("data"), packet.Data); err != nil {
		return err
	}

	if len(packet.ExpireAt) > 0 {
		if err = pBuck.Put([]byte("expireAt"), []byte(packet.ExpireAt)); err != nil {
			return err
		}
	}

	return nil
}
