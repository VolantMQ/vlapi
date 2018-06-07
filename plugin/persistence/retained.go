package persistence

type retained struct {
	status  *dbStatus
	packets PersistedPackets
}

func (r *retained) Load() (PersistedPackets, error) {
	return r.packets, nil
}

func (r *retained) Store(data PersistedPackets) error {
	r.packets = data

	return nil
}

func (r *retained) Wipe() error {
	r.packets = PersistedPackets{}
	return nil
}
