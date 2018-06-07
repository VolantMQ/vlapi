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
	lock    sync.Mutex
	state   *SessionState
	packets PersistedPackets
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

func (s *sessions) PacketsForEach(id []byte, loader PacketLoader) error {
	if elem, ok := s.entries.Load(string(id)); ok {
		ses := elem.(*session)

		for _, p := range ses.packets {
			loader.LoadPersistedPacket(p) // nolint: errcheck
		}
	}

	return nil
}

func (s *sessions) PacketsStore(id []byte, packets PersistedPackets) error {
	elem, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	ses := elem.(*session)

	ses.lock.Lock()
	ses.packets = append(ses.packets, packets...)
	ses.lock.Unlock()

	return nil
}

func (s *sessions) PacketsDelete(id []byte) error {
	if elem, ok := s.entries.Load(string(id)); ok {
		ses := elem.(*session)

		ses.lock.Lock()
		ses.packets = PersistedPackets{}
		ses.lock.Unlock()
	}

	return nil
}

func (s *sessions) PacketStore(id []byte, pkt *PersistedPacket) error {
	elem, loaded := s.entries.LoadOrStore(string(id), &session{})
	if !loaded {
		atomic.AddUint64(&s.count, 1)
	}

	ses := elem.(*session)

	ses.lock.Lock()
	ses.packets = append(ses.packets, pkt)
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
