package vlplugin

import (
	"errors"

	"net/http"

	"github.com/VolantMQ/vlapi/plugin/health"
	"github.com/VolantMQ/vlapi/subscriber"
	"go.uber.org/zap"
)

// APIVersion version of current API
const APIVersion = "1.0.0"

var (
	// ErrInvalidArgs invalid arguments
	ErrInvalidArgs = errors.New("plugin: invalid arguments")
)

// Descriptor describes plugin
type Descriptor struct {
	V string
	N string
	D string
	T string
}

// Info return plugin information
type Info interface {
	// Version in format major.minor.patch
	Version() (string, string)
	// Name plugin name
	Name() string
	// Desc plugin description
	Desc() string
	// Type plugin type
	Type() string
}

// nolint: golint
type Messaging interface {
	GetSubscriber(id string) (vlsubscriber.IFace, error)
}

// HTTPHandler provided by VolantMQ server
type HTTPHandler interface {
	Mux() *http.ServeMux
	Addr() string
}

// HTTP ...
type HTTP interface {
	GetHTTPServer(port string) HTTPHandler
}

type Health interface {
	GetHealth() vlhealth.Handler
}

// SysParams system-wide config passed to plugin
type SysParams struct {
	Messaging
	HTTP
	Health
	Log           *zap.SugaredLogger
	SignalFailure func(name, msg string)
}

// Plugin entry to plugin
type Plugin interface {
	// Init initialize plugin
	// might accepts interface which specifies config
	// return interface to plugin entry
	Load(interface{}, *SysParams) (interface{}, error)

	// Info plugin information
	Info() Info
}

// nolint: golint
type Must interface {
	Shutdown() error
}

// Version of plugin. Version format format major.minor.patch
// returns API version plugin is built with
//         plugin version
func (b *Descriptor) Version() (string, string) {
	return APIVersion, b.V
}

// Name of plugin
func (b *Descriptor) Name() string {
	return b.N
}

// Desc of plugin
func (b *Descriptor) Desc() string {
	return b.D
}

// Type of plugin
func (b *Descriptor) Type() string {
	return b.T
}
