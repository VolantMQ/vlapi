package persistenceMem

import (
	"github.com/VolantMQ/vlapi/plugin/persistence"
)

type retained struct {
	status  *dbStatus
	packets []*persistence.PersistedPacket
}

func (r *retained) Load() ([]*persistence.PersistedPacket, error) {
	return r.packets, nil
}

func (r *retained) Store(data []*persistence.PersistedPacket) error {
	r.packets = data

	return nil
}

func (r *retained) Wipe() error {
	r.packets = []*persistence.PersistedPacket{}
	return nil
}
