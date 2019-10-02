package main

import (
	"errors"
	"fmt"
	"net/http/pprof"
	"reflect"
	"runtime"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/VolantMQ/vlapi/vlplugin"
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
	Plugin.V = vlplugin.Version()
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
	Block   int    `mapstructure:"block,omitempty" yaml:"block,omitempty" json:"block,omitempty" default:"0"`
	Mutex   int    `mapstructure:"mutex,omitempty" yaml:"mutex,omitempty" json:"mutex,omitempty" default:"0"`
}

type impl struct {
	*vlplugin.SysParams
	cfg config
}

func (pl *pl) Load(c interface{}, params *vlplugin.SysParams) (pla interface{}, err error) {
	p := &impl{
		SysParams: params,
	}

	decodeIface := func() error {
		var data []byte
		var e error
		if data, e = yaml.Marshal(c); e != nil {
			e = errors.New(Plugin.T + "." + Plugin.N + ": " + e.Error())
			return e
		}

		if e = yaml.Unmarshal(data, &p.cfg); e != nil {
			e = errors.New(Plugin.T + "." + Plugin.N + ": " + e.Error())
			return e
		}

		return e
	}

	switch c.(type) {
	case map[string]interface{}:
		if err = decodeIface(); err != nil {
			return
		}
	case map[interface{}]interface{}:
		if err = decodeIface(); err != nil {
			return
		}
	case []byte:
		if err = yaml.Unmarshal(c.([]byte), &p.cfg); err != nil {
			err = errors.New(Plugin.T + "." + Plugin.N + ": " + err.Error())
			return
		}
	default:
		err = fmt.Errorf("%s.%s: invalid config type %s", Plugin.T, Plugin.N, reflect.TypeOf(c).String())
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
		params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc(p.cfg.Path+"/profile", pprof.Profile)
	}

	if p.cfg.Symbol {
		params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc(p.cfg.Path+"/symbol", pprof.Symbol)
	}

	if p.cfg.Trace {
		params.GetHTTPServer(p.cfg.Port).Mux().HandleFunc(p.cfg.Path+"/trace", pprof.Trace)
	}

	if p.cfg.Block > 0 {
		runtime.SetBlockProfileRate(p.cfg.Block)
	}

	if p.cfg.Mutex > 0 {
		runtime.SetBlockProfileRate(p.cfg.Mutex)
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
