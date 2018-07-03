package persistence

import (
	"sync"
	"sync/atomic"
)

type sessions struct {
	status  *dbStatus
	entries sync.Map
	count   uint64
}

type session struct {
	lock  sync.Mutex
	state *SessionState
	PersistedPackets
}

var _ Sessions = (*sessions)(nil)

func (s *sessions) Exists(id []byte) bool {
	_, ok := s.entries.Load(string(id))
	return ok
}

func (s *sessions) Count() uint64 {
	return atomic.LoadUint64(&s.count)
}

func (s *sessions) SubscriptionsStore(id []byte, data []byte) error {
	elem, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	ses := elem.(*session)
	ses.lock.Lock()
	defer ses.lock.Unlock()
	if ses.state == nil {
		ses.state = &SessionState{}
	}

	ses.state.Subscriptions = data

	return nil
}

func (s *sessions) SubscriptionsDelete(id []byte) error {
	if elem, ok := s.entries.Load(string(id)); ok {
		ses := elem.(*session)
		ses.lock.Lock()
		defer ses.lock.Unlock()
		ses.state.Subscriptions = []byte{}
	}

	return nil
}

func (s *sessions) PacketsForEachQoS0(id []byte, ctx interface{}, load PacketLoader) error {
	if elem, ok := s.entries.Load(string(id)); ok {
		ses := elem.(*session)

		ses.lock.Lock()
		for i := len(ses.QoS0) - 1; i >= 0; i-- {
			rm, err := load(ctx, ses.QoS0[i])

			if rm {
				ses.QoS0 = append(ses.QoS0[:i], ses.QoS0[i+1:]...)
			}

			if err != nil {
				break
			}
		}

		ses.lock.Unlock()
	}

	return nil
}

func (s *sessions) PacketsForEachQoS12(id []byte, ctx interface{}, load PacketLoader) error {
	if elem, ok := s.entries.Load(string(id)); ok {
		ses := elem.(*session)

		ses.lock.Lock()
		for i := len(ses.QoS12) - 1; i >= 0; i-- {
			rm, err := load(ctx, ses.QoS12[i])

			if rm {
				ses.QoS12 = append(ses.QoS12[:i], ses.QoS12[i+1:]...)
			}

			if err != nil {
				break
			}
		}

		ses.lock.Unlock()
	}

	return nil
}

func (s *sessions) PacketsForEachUnAck(id []byte, ctx interface{}, load PacketLoader) error {
	if elem, ok := s.entries.Load(string(id)); ok {
		ses := elem.(*session)

		ses.lock.Lock()

		for i := len(ses.UnAck) - 1; i >= 0; i-- {
			rm, err := load(ctx, ses.UnAck[i])

			if rm {
				ses.UnAck = append(ses.UnAck[:i], ses.UnAck[i+1:]...)
			}

			if err != nil {
				break
			}
		}

		ses.lock.Unlock()
	}

	return nil
}

func (s *sessions) PacketCountQoS0(id []byte) (int, error) {
	elem, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	ses := elem.(*session)

	defer ses.lock.Unlock()
	ses.lock.Lock()

	return len(ses.QoS0), nil
}

func (s *sessions) PacketCountQoS12(id []byte) (int, error) {
	elem, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	ses := elem.(*session)

	defer ses.lock.Unlock()
	ses.lock.Lock()

	return len(ses.QoS12), nil
}

func (s *sessions) PacketCountUnAck(id []byte) (int, error) {
	elem, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	ses := elem.(*session)

	defer ses.lock.Unlock()
	ses.lock.Lock()

	return len(ses.UnAck), nil
}

func (s *sessions) PacketsStore(id []byte, packets PersistedPackets) error {
	elem, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	ses := elem.(*session)

	ses.lock.Lock()
	ses.QoS0 = append(ses.QoS0, packets.QoS0...)
	ses.QoS12 = append(ses.QoS12, packets.QoS12...)
	ses.UnAck = append(ses.UnAck, packets.UnAck...)
	ses.lock.Unlock()

	return nil
}

func (s *sessions) PacketsDelete(id []byte) error {
	if elem, ok := s.entries.Load(string(id)); ok {
		ses := elem.(*session)

		ses.lock.Lock()
		ses.QoS0 = []*PersistedPacket{}
		ses.QoS12 = []*PersistedPacket{}
		ses.UnAck = []*PersistedPacket{}
		ses.lock.Unlock()
	}

	return nil
}

func (s *sessions) PacketStoreQoS0(id []byte, pkt *PersistedPacket) error {
	elem, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	ses := elem.(*session)

	ses.lock.Lock()
	ses.QoS0 = append(ses.QoS0, pkt)
	ses.lock.Unlock()

	return nil
}

func (s *sessions) PacketStoreQoS12(id []byte, pkt *PersistedPacket) error {
	elem, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	ses := elem.(*session)

	ses.lock.Lock()
	ses.QoS12 = append(ses.QoS12, pkt)
	ses.lock.Unlock()

	return nil
}

func (s *sessions) PacketStoreUnAck(id []byte, pkt *PersistedPacket) error {
	elem, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	ses := elem.(*session)

	ses.lock.Lock()
	ses.UnAck = append(ses.UnAck, pkt)
	ses.lock.Unlock()

	return nil
}

func (s *sessions) LoadForEach(loader SessionLoader, context interface{}) error {
	var err error
	s.entries.Range(func(key, value interface{}) bool {
		sID := []byte(key.(string))
		ses := value.(*SessionState)
		err = loader.LoadSession(context, sID, ses)
		return err == nil
	})

	return err
}

func (s *sessions) Create(id []byte, state *SessionBase) error {
	_, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	return nil
}

func (s *sessions) StateStore(id []byte, state *SessionState) error {
	elem, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	ses := elem.(*session)
	ses.lock.Lock()
	defer ses.lock.Unlock()

	if ses.state != nil {
		if len(ses.state.Subscriptions) > 0 && len(state.Subscriptions) == 0 {
			state.Subscriptions = ses.state.Subscriptions
		}
	}

	ses.state = state
	return nil
}

func (s *sessions) ExpiryStore(id []byte, exp *SessionDelays) error {
	elem, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	ses := elem.(*session)
	ses.lock.Lock()
	defer ses.lock.Unlock()

	ses.state.Expire = exp
	return nil
}

func (s *sessions) ExpiryDelete(id []byte) error {
	elem, loaded := s.entries.Load(string(id))
	if loaded {
		ses := elem.(*session)
		ses.lock.Lock()
		defer ses.lock.Unlock()

		ses.state.Expire = nil
	}

	return nil
}

func (s *sessions) StateDelete(id []byte) error {
	if elem, ok := s.entries.Load(string(id)); ok {
		ses := elem.(*session)
		ses.lock.Lock()
		defer ses.lock.Unlock()
		ses.state = nil
	}

	return nil
}

// Delete
func (s *sessions) Delete(id []byte) error {
	s.entries.Delete(string(id))
	atomic.AddUint64(&s.count, ^uint64(0))
	return nil
}
