package main

import (
	"errors"
	"net/http/pprof"

	"strings"

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

const (
	defaultPath = "/debug/pprof"
)

func init() {
	Plugin.V = "0.0.1"
	Plugin.N = "prof.profiler"
	Plugin.T = "debug"
}

type config struct {
	Port    string `mapstructure:"port,omitempty" yaml:"port,omitempty" json:"port,omitempty" default:""`
	Path    string `mapstructure:"path,omitempty" yaml:"path,omitempty" json:"path,omitempty" default:""`
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

	if p.cfg.Path == "" {
		p.cfg.Path = defaultPath
	}

	p.cfg.Path = strings.TrimSuffix(p.cfg.Path, "/")

	params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc(p.cfg.Path+"/", pprof.Index)
	if p.cfg.CmdLine {
		params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc(p.cfg.Path+"/cmdline", pprof.Cmdline)
	}

	if p.cfg.CPU {
		params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc(p.cfg.Path+"profile", pprof.Profile)
	}

	if p.cfg.Symbol {
		params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc(p.cfg.Path+"/symbol", pprof.Symbol)
	}

	if p.cfg.Trace {
		params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc(p.cfg.Path+"/trace", pprof.Trace)
	}

	pla = p

	p.Log.Infof("profiler available at [http://%s%s]", params.GetHTTPServer(p.cfg.Port).Addr(), p.cfg.Path)
	return
}

// Info plugin info
func (pl *pl) Info() vlplugin.Info {
	return pl
}

func main() {
	panic("this is a plugin, build it as with -buildmode=plugin")
}
