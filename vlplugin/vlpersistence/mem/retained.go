package persistenceMem

import "github.com/VolantMQ/vlapi/vlplugin/vlpersistence"

type retained struct {
	status  *dbStatus
	packets []*vlpersistence.PersistedPacket
}

func (r *retained) Load() ([]*vlpersistence.PersistedPacket, error) {
	return r.packets, nil
}

func (r *retained) Store(data []*vlpersistence.PersistedPacket) error {
	r.packets = data

	return nil
}

func (r *retained) Wipe() error {
	r.packets = []*vlpersistence.PersistedPacket{}
	return nil
}
