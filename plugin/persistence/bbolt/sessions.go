package persistenceBbolt

import (
	"encoding/binary"
	"errors"

	"github.com/VolantMQ/vlapi/plugin/persistence"
	"github.com/etcd-io/bbolt"
)

type sessions struct {
	*dbStatus
}

var _ persistence.Sessions = (*sessions)(nil)

func (s *sessions) init() error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		val := sessions.Get(sessionsCount)
		if len(val) == 0 {
			buf := [8]byte{}
			num := binary.PutUvarint(buf[:], 0)
			if err := sessions.Put(sessionsCount, buf[:num]); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *sessions) Exists(id []byte) bool {
	err := s.db.View(func(tx *bbolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		if ses := sessions.Bucket(id); ses == nil {
			return bbolt.ErrBucketNotFound
		}

		return nil
	})

	return err == nil
}

func (s *sessions) Count() uint64 {
	var count uint64

	s.db.View(func(tx *bbolt.Tx) error { // nolint: errcheck, gosec
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
	return s.db.Update(func(tx *bbolt.Tx) error {
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
	return s.db.Update(func(tx *bbolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		session := tx.Bucket(id)
		if session == nil {
			return persistence.ErrNotFound
		}

		if err := session.Delete(bucketSubscriptions); err != nil && err != bbolt.ErrBucketNotFound {
			return err
		}
		return nil
	})
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

func (s *sessions) PacketCountQoS0(id []byte) (uint64, error) {
	return s.packetCount(id, bucketPacketsQoS0)
}

func (s *sessions) PacketCountQoS12(id []byte) (uint64, error) {
	return s.packetCount(id, bucketPacketsQoS12)
}

func (s *sessions) PacketCountUnAck(id []byte) (uint64, error) {
	return s.packetCount(id, bucketPacketsUnAck)
}

func (s *sessions) PacketsStore(id []byte, packets persistence.PersistedPackets) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
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
	return s.db.Update(func(tx *bbolt.Tx) error {
		return s.packetsStore(tx, id, bucketPacketsQoS0, []*persistence.PersistedPacket{pkt})
	})
}

func (s *sessions) PacketStoreQoS12(id []byte, pkt *persistence.PersistedPacket) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		return s.packetsStore(tx, id, bucketPacketsQoS12, []*persistence.PersistedPacket{pkt})
	})
}

func (s *sessions) PacketStoreUnAck(id []byte, pkt *persistence.PersistedPacket) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		return s.packetsStore(tx, id, bucketPacketsUnAck, []*persistence.PersistedPacket{pkt})
	})
}

func (s *sessions) PacketsDelete(id []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		if sessions := tx.Bucket(bucketSessions); sessions != nil {
			ses := sessions.Bucket(id)
			if ses == nil {
				return persistence.ErrNotFound
			}
			if err := ses.DeleteBucket(bucketPackets); err != nil && err != bbolt.ErrBucketNotFound {
				return err
			}
		}
		return nil
	})
}

func (s *sessions) LoadForEach(loader persistence.SessionLoader, context interface{}) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
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
	return s.db.Update(func(tx *bbolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		session, err := getSession(id, sessions)
		if err != nil {
			return err
		}

		var st *bbolt.Bucket
		if st, err = session.CreateBucketIfNotExists(bucketState); err != nil {
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

func (s *sessions) Delete(id []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
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

		return sessions.Put(sessionsCount, buf[:num])
	})
}

func (s *sessions) StateStore(id []byte, state *persistence.SessionState) error {
	if state.Expire != nil && len(state.Expire.ExpireIn) == 0 {
		return persistence.ErrInvalidArgs
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		session, err := getSession(id, sessions)
		if err != nil {
			return err
		}

		var st *bbolt.Bucket
		if st, err = session.CreateBucketIfNotExists(bucketState); err != nil {
			return err
		}

		if len(state.Subscriptions) > 0 {
			if err = st.Put(bucketSubscriptions, state.Subscriptions); err != nil {
				return err
			}
		}

		if state.Expire != nil {
			err = expiryPut(st, state.Expire)
		}

		return err
	})
}

func (s *sessions) ExpiryStore(id []byte, exp *persistence.SessionDelays) error {
	if exp == nil || len(exp.ExpireIn) == 0 {
		return persistence.ErrInvalidArgs
	}

	return s.db.Update(func(tx *bbolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		session, err := getSession(id, sessions)
		if err != nil {
			return err
		}

		var st *bbolt.Bucket
		st, err = session.CreateBucketIfNotExists(bucketState)
		if err != nil {
			return err
		}

		return expiryPut(st, exp)
	})
}

func (s *sessions) ExpiryDelete(id []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
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

		if err := state.DeleteBucket(bucketExpire); err != nil && err != bbolt.ErrBucketNotFound {
			return err
		}

		return nil
	})
}

func (s *sessions) StateDelete(id []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
		sessions := tx.Bucket(bucketSessions)
		if sessions == nil {
			return persistence.ErrNotInitialized
		}

		session, err := getSession(id, sessions)
		if err != nil {
			return err
		}

		if err := session.DeleteBucket(bucketState); err != nil && err != bbolt.ErrBucketNotFound {
			return err
		}

		return nil
	})
}

func expiryPut(state *bbolt.Bucket, exp *persistence.SessionDelays) error {
	expire, err := state.CreateBucketIfNotExists(bucketExpire)
	if err != nil {
		return err
	}

	if err = expire.Put([]byte("since"), []byte(exp.Since)); err != nil {
		return err
	}

	if err = expire.Put([]byte("expireIn"), []byte(exp.ExpireIn)); err != nil {
		return err
	}

	if len(exp.Will) > 0 {
		if err = expire.Put([]byte("will"), exp.Will); err != nil {
			return err
		}
	}

	return nil
}

func (s *sessions) packetsForEach(id []byte, ctx interface{}, load persistence.PacketLoader, bucket []byte) error {
	return s.db.Update(func(tx *bbolt.Tx) error {
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

		return packets.ForEach(func(k, v []byte) error {
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
				if err != nil {
					return err
				}

				if rm {
					if err = packets.DeleteBucket(k); err != nil && err != bbolt.ErrBucketNotFound {
						return err
					}
				}
			}
			return nil
		})
	})
}

func (s *sessions) packetCount(id []byte, bucket []byte) (uint64, error) {
	var count uint64

	err := s.db.View(func(tx *bbolt.Tx) error {
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

		count = uint64(packets.Stats().InlineBucketN)

		return nil
	})

	return count, err
}

func (s *sessions) packetsStore(tx *bbolt.Tx, id []byte, bucket []byte, packets []*persistence.PersistedPacket) error {
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

func getSession(id []byte, sessions *bbolt.Bucket) (*bbolt.Bucket, error) {
	session := sessions.Bucket(id)
	if session == nil {
		var err error
		if session, err = sessions.CreateBucket(id); err != nil {
			return nil, err
		}

		count, num := binary.Uvarint(sessions.Get(sessionsCount))
		if num <= 0 {
			if err = sessions.DeleteBucket(id); err != nil && err != bbolt.ErrBucketNotFound {
				return nil, err
			}
			return nil, nil
		}

		count++

		buf := [8]byte{}
		num = binary.PutUvarint(buf[:], count)
		if err = sessions.Put([]byte("count"), buf[:num]); err != nil {
			return nil, err
		}
	}

	return session, nil
}

func createPacketsBucket(tx *bbolt.Tx, id []byte, bucket []byte) (*bbolt.Bucket, error) {
	sessions, err := tx.CreateBucketIfNotExists(bucketSessions)
	if err != nil {
		return nil, err
	}

	var session *bbolt.Bucket
	if session, err = getSession(id, sessions); err != nil {
		return nil, err
	}

	var packetsRoot *bbolt.Bucket
	if packetsRoot, err = session.CreateBucketIfNotExists(bucketPackets); err != nil {
		return nil, err
	}

	return packetsRoot.CreateBucketIfNotExists(bucket)
}

func storePacket(buck *bbolt.Bucket, packet *persistence.PersistedPacket) error {
	id, err := buck.NextSequence()
	if err != nil {
		return nil
	}

	pBuck, e := buck.CreateBucketIfNotExists(itob64(id))
	if e != nil {
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
