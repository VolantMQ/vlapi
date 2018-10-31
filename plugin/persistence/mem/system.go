package persistenceMem

import (
	"github.com/VolantMQ/vlapi/plugin/persistence"
)

type system struct {
	status *dbStatus
}

func (s *system) GetInfo() (*persistence.SystemState, error) {
	state := &persistence.SystemState{}

	return state, nil
}

func (s *system) SetInfo(state *persistence.SystemState) error {
	return nil
}
