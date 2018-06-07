package main

import (
	"errors"
	"net/http/pprof"

	"github.com/VolantMQ/vlapi/plugin"
	"gopkg.in/yaml.v2"
)

type pl struct {
	vlplugin.Descriptor
}

var _ vlplugin.Plugin = (*pl)(nil)
var _ vlplugin.Info = (*pl)(nil)

// Plugin symbol
var Plugin pl

func init() {
	Plugin.V = "0.0.1"
	Plugin.N = "prof.profiler"
	Plugin.T = "debug"
}

type config struct {
	Port    string `mapstructure:"port,omitempty" yaml:"port,omitempty" json:"port,omitempty" default:""`
	CPU     bool   `mapstructure:"cpu,omitempty" yaml:"cpu,omitempty" json:"cpu,omitempty" default:"false"`
	Memory  bool   `mapstructure:"memory,omitempty" yaml:"memory,omitempty" json:"memory,omitempty" default:"false"`
	Trace   bool   `mapstructure:"trace,omitempty" yaml:"trace,omitempty" json:"trace,omitempty" default:"false"`
	Symbol  bool   `mapstructure:"symbol,omitempty" yaml:"symbol,omitempty" json:"symbol,omitempty" default:"false"`
	CmdLine bool   `mapstructure:"cmdline,omitempty" yaml:"cmdline,omitempty" json:"cmdline,omitempty" default:"false"`
}

type impl struct {
	*vlplugin.SysParams
	cfg config
}

func (pl *pl) Load(c interface{}, params *vlplugin.SysParams) (pla interface{}, err error) {
	p := &impl{
		SysParams: params,
	}

	switch c.(type) {
	case map[interface{}]interface{}:
		var data []byte
		if data, err = yaml.Marshal(c); err != nil {
			err = errors.New(Plugin.T + "." + Plugin.N + ": " + err.Error())
			return
		}

		if err = yaml.Unmarshal(data, &p.cfg); err != nil {
			err = errors.New(Plugin.T + "." + Plugin.N + ": " + err.Error())
			return
		}
	case []byte:
		if err = yaml.Unmarshal(c.([]byte), &p.cfg); err != nil {
			err = errors.New(Plugin.T + "." + Plugin.N + ": " + err.Error())
			return
		}
	default:
		err = errors.New(Plugin.T + "." + Plugin.N + ": invalid config")
		return
	}

	params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc("/debug/pprof/", pprof.Index)
	if p.cfg.CmdLine {
		params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
	}

	if p.cfg.CPU {
		params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc("/debug/pprof/profile", pprof.Profile)
	}

	if p.cfg.Symbol {
		params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc("/debug/pprof/symbol", pprof.Symbol)
	}

	if p.cfg.Trace {
		params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc("/debug/pprof/trace", pprof.Trace)
	}

	pla = p

	p.Log.Infof("profiler available at [http://%s/debug/pprof]", params.GetHTTPServer(p.cfg.Port).Addr())
	return
}

// Info plugin info
func (pl *pl) Info() vlplugin.Info {
	return pl
}

func main() {
	panic("this is a plugin, build it as with -buildmode=plugin")
}
