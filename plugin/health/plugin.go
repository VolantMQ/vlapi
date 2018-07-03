package health

import (
	"errors"
	"strings"

	"github.com/VolantMQ/vlapi/plugin"
	"github.com/troian/healthcheck"
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
	defaultPath = "/health"
)

func init() {
	Plugin.V = "0.0.1"
	Plugin.N = "health"
	Plugin.T = "health"
}

type config struct {
	Port              string `mapstructure:"port,omitempty" yaml:"port,omitempty" json:"port,omitempty" default:""`
	Path              string `mapstructure:"path,omitempty" yaml:"path,omitempty" json:"path,omitempty" default:""`
	LivenessEndpoint  string `mapstructure:"livenessEndpoint,omitempty" yaml:"livenessEndpoint,omitempty" json:"livenessEndpoint,omitempty" default:""`
	ReadinessEndpoint string `mapstructure:"readinessEndpoint,omitempty" yaml:"readinessEndpoint,omitempty" readinessEndpoint:"path,omitempty" default:""`
}

type impl struct {
	*vlplugin.SysParams
	healthcheck.Handler
	cfg config
}

var _ healthcheck.Handler = (*impl)(nil)

func (pl *pl) Load(c interface{}, params *vlplugin.SysParams) (pla interface{}, err error) {
	p := &impl{
		SysParams: params,
		Handler:   healthcheck.NewHandler(),
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

	mux := params.GetHTTPServer(p.cfg.Port).Mux()

	mux.HandleFunc(p.cfg.Path+"/"+p.cfg.LivenessEndpoint, p.LiveEndpoint)
	mux.HandleFunc(p.cfg.Path+"/"+p.cfg.ReadinessEndpoint, p.ReadyEndpoint)

	pla = p

	p.Log.Infof("health check liveness endpoint [http://%s%s/%s]", params.GetHTTPServer(p.cfg.Port).Addr(), p.cfg.Path, p.cfg.LivenessEndpoint)
	p.Log.Infof("health check readiness endpoint [http://%s%s/%s]", params.GetHTTPServer(p.cfg.Port).Addr(), p.cfg.Path, p.cfg.ReadinessEndpoint)
	return
}

// Info plugin info
func (pl *pl) Info() vlplugin.Info {
	return pl
}

func main() {
	panic("this is a plugin, build it as with -buildmode=plugin")
}
