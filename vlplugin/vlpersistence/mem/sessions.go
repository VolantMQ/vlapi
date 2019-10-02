package persistenceMem

import (
	"sync"

	"github.com/VolantMQ/vlapi/vlplugin/vlpersistence"
)

type sessions struct {
	status  *dbStatus
	entries map[string]*session
	lock    sync.RWMutex
}

type session struct {
	state *vlpersistence.SessionState
	vlpersistence.PersistedPackets
}

var _ vlpersistence.Sessions = (*sessions)(nil)

func (s *sessions) Exists(id []byte) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()

	_, ok := s.entries[string(id)]
	return ok
}

func (s *sessions) Count() uint64 {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return uint64(len(s.entries))
}

func (s *sessions) SubscriptionsStore(id []byte, data []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	ses, loaded := s.entries[string(id)]
	if !loaded {
		ses = &session{}
		s.entries[string(id)] = ses
	}

	if ses.state == nil {
		ses.state = &vlpersistence.SessionState{}
	}

	ses.state.Subscriptions = data

	return nil
}

func (s *sessions) SubscriptionsDelete(id []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	ses, loaded := s.entries[string(id)]
	if loaded {
		ses.state.Subscriptions = []byte{}
	}

	return nil
}

func (s *sessions) PacketsForEachQoS0(id []byte, ctx interface{}, load vlpersistence.PacketLoader) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		for i := len(ses.QoS0) - 1; i >= 0; i-- {
			rm, err := load(ctx, ses.QoS0[i])

			if rm {
				ses.QoS0 = append(ses.QoS0[:i], ses.QoS0[i+1:]...)
			}

			if err != nil {
				break
			}
		}
	}

	return nil
}

func (s *sessions) PacketsForEachQoS12(id []byte, ctx interface{}, load vlpersistence.PacketLoader) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		for i := len(ses.QoS12) - 1; i >= 0; i-- {
			rm, err := load(ctx, ses.QoS12[i])

			if rm {
				ses.QoS12 = append(ses.QoS12[:i], ses.QoS12[i+1:]...)
			}

			if err != nil {
				break
			}
		}
	}

	return nil
}

func (s *sessions) PacketsForEachUnAck(id []byte, ctx interface{}, load vlpersistence.PacketLoader) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		for i := len(ses.UnAck) - 1; i >= 0; i-- {
			rm, err := load(ctx, ses.UnAck[i])

			if rm {
				ses.UnAck = append(ses.UnAck[:i], ses.UnAck[i+1:]...)
			}

			if err != nil {
				break
			}
		}
	}

	return nil
}

func (s *sessions) PacketCountQoS0(id []byte) (uint64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		return uint64(len(ses.QoS0)), nil
	}

	return 0, vlpersistence.ErrNotFound
}

func (s *sessions) PacketCountQoS12(id []byte) (uint64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		return uint64(len(ses.QoS12)), nil
	}

	return 0, vlpersistence.ErrNotFound
}

func (s *sessions) PacketCountUnAck(id []byte) (uint64, error) {
	s.lock.RLock()
	defer s.lock.RUnlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		return uint64(len(ses.UnAck)), nil
	}

	return 0, vlpersistence.ErrNotFound
}

func (s *sessions) PacketsStore(id []byte, packets vlpersistence.PersistedPackets) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		ses.QoS0 = append(ses.QoS0, packets.QoS0...)
		ses.QoS12 = append(ses.QoS12, packets.QoS12...)
		ses.UnAck = append(ses.UnAck, packets.UnAck...)

		return nil
	}

	return vlpersistence.ErrNotFound
}

func (s *sessions) PacketsDelete(id []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		ses.QoS0 = []*vlpersistence.PersistedPacket{}
		ses.QoS12 = []*vlpersistence.PersistedPacket{}
		ses.UnAck = []*vlpersistence.PersistedPacket{}

		return nil
	}

	return vlpersistence.ErrNotFound
}

func (s *sessions) PacketStoreQoS0(id []byte, pkt *vlpersistence.PersistedPacket) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		ses.QoS0 = append(ses.QoS0, pkt)
		return nil
	}

	return vlpersistence.ErrNotFound
}

func (s *sessions) PacketStoreQoS12(id []byte, pkt *vlpersistence.PersistedPacket) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		ses.QoS12 = append(ses.QoS12, pkt)
		return nil
	}

	return vlpersistence.ErrNotFound
}

func (s *sessions) PacketStoreUnAck(id []byte, pkt *vlpersistence.PersistedPacket) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		ses.UnAck = append(ses.UnAck, pkt)
		return nil
	}

	return vlpersistence.ErrNotFound
}

func (s *sessions) LoadForEach(loader vlpersistence.SessionLoader, context interface{}) error {
	s.lock.RLock()
	defer s.lock.RUnlock()

	var err error

	for id, ses := range s.entries {
		if err = loader.LoadSession(context, []byte(id), ses.state); err != nil {
			break
		}
	}

	return err
}

func (s *sessions) Create(id []byte, state *vlpersistence.SessionBase) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, loaded := s.entries[string(id)]; !loaded {
		s.entries[string(id)] = &session{}
		return nil
	}

	return vlpersistence.ErrAlreadyExists
}

func (s *sessions) StateStore(id []byte, state *vlpersistence.SessionState) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		if ses.state != nil {
			if len(ses.state.Subscriptions) > 0 && len(state.Subscriptions) == 0 {
				state.Subscriptions = ses.state.Subscriptions
			}
		}

		ses.state = state
		return nil
	}

	return vlpersistence.ErrNotFound
}

func (s *sessions) ExpiryStore(id []byte, exp *vlpersistence.SessionDelays) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		if ses.state == nil {
			ses.state = &vlpersistence.SessionState{}
		}
		ses.state.Expire = exp

		return nil
	}

	return vlpersistence.ErrNotFound
}

func (s *sessions) ExpiryDelete(id []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		ses.state.Expire = nil

		return nil
	}

	return vlpersistence.ErrNotFound
}

func (s *sessions) StateDelete(id []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if ses, loaded := s.entries[string(id)]; loaded {
		ses.state = nil

		return nil
	}

	return vlpersistence.ErrNotFound
}

func (s *sessions) Delete(id []byte) error {
	s.lock.Lock()
	defer s.lock.Unlock()

	if _, loaded := s.entries[string(id)]; loaded {
		delete(s.entries, string(id))

		return nil
	}

	return vlpersistence.ErrNotFound
}
