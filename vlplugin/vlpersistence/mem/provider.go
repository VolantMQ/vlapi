package persistenceMem

import (
	"github.com/VolantMQ/vlapi/vlplugin"
	"github.com/VolantMQ/vlapi/vlplugin/vlpersistence"
)

type dbStatus struct {
	done chan struct{}
}

type impl struct {
	status dbStatus
	r      retained
	s      sessions
	sys    system
}

// nolint: golint
func Load(c interface{}, params *vlplugin.SysParams) (vlpersistence.IFace, error) {
	pl := &impl{}

	pl.status.done = make(chan struct{})

	pl.r = retained{
		status: &pl.status,
	}

	pl.s = sessions{
		status:  &pl.status,
		entries: make(map[string]*session),
	}

	pl.sys = system{
		status: &pl.status,
	}

	return pl, nil
}

func (p *impl) System() (vlpersistence.System, error) {
	select {
	case <-p.status.done:
		return nil, vlpersistence.ErrNotOpen
	default:
	}

	return &p.sys, nil
}

func (p *impl) Sessions() (vlpersistence.Sessions, error) {
	select {
	case <-p.status.done:
		return nil, vlpersistence.ErrNotOpen
	default:
	}

	return &p.s, nil
}

func (p *impl) Retained() (vlpersistence.Retained, error) {
	select {
	case <-p.status.done:
		return nil, vlpersistence.ErrNotOpen
	default:
	}

	return &p.r, nil
}

func (p *impl) Shutdown() error {
	select {
	case <-p.status.done:
		return vlpersistence.ErrNotOpen
	default:
		close(p.status.done)
	}

	return nil
}
