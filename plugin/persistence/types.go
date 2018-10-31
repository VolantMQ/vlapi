package persistence

// Errors persistence errors
type Errors int

const (
	// ErrInvalidArgs invalid arguments provided
	ErrInvalidArgs Errors = iota
	// ErrUnknownProvider if provider is unknown
	ErrUnknownProvider
	// ErrInvalidConfig provided config is invalid
	ErrInvalidConfig
	// ErrAlreadyExists object already exists
	ErrAlreadyExists
	// ErrNotInitialized persistence provider not initialized yet
	ErrNotInitialized
	// ErrNotFound object not found
	ErrNotFound
	// ErrNotOpen storage is not open
	ErrNotOpen
	// ErrBrokenEntry persisted entry does not meet requirements
	ErrBrokenEntry
)

var errorsDesc = map[Errors]string{
	ErrInvalidArgs:     "persistence: invalid arguments",
	ErrUnknownProvider: "persistence: unknown provider",
	ErrInvalidConfig:   "persistence: invalid config",
	ErrAlreadyExists:   "persistence: already exists",
	ErrNotInitialized:  "persistence: not initialized",
	ErrNotFound:        "persistence: not found",
	ErrNotOpen:         "persistence: not open",
	ErrBrokenEntry:     "persistence: broken entry",
}

// Errors description during persistence
func (e Errors) Error() string {
	if s, ok := errorsDesc[e]; ok {
		return s
	}

	return "persistence: unknown error"
}

// PersistedPacket wraps packet to handle misc cases like expiration
type PersistedPacket struct {
	// ExpireAt shows either packet has expiration value
	ExpireAt string
	// Data is encoded byte stream as it goes over network
	Data []byte
}

// PersistedPackets array of persisted packets
type PersistedPackets struct {
	QoS0  []*PersistedPacket
	QoS12 []*PersistedPacket
	UnAck []*PersistedPacket
}

// SessionDelays formerly known as expiry set timestamp to handle will delay and/or expiration
type SessionDelays struct {
	Since    string
	ExpireIn string
	Will     []byte
}

// SessionBase ...
type SessionBase struct {
	Timestamp string
	Version   byte
}

// SessionState object
type SessionState struct {
	Subscriptions []byte
	Errors        []error
	Expire        *SessionDelays
	SessionBase
}

// SystemState system configuration
type SystemState struct {
	Version   string
	CreatedAt string
}

// PacketLoader application callback doing packed decode
// when return true in first return packet will be deleted
// if error presented load interrupted after current packet
type PacketLoader func(interface{}, *PersistedPacket) (bool, error)

// SessionLoader implemented by session manager to load persisted sessions when server starts
type SessionLoader interface {
	LoadSession(ctx interface{}, id []byte, state *SessionState) error
}

// Packets interface for connection to handle packets
type Packets interface {
	PacketCountQoS0(id []byte) (uint64, error)
	PacketCountQoS12(id []byte) (uint64, error)
	PacketCountUnAck(id []byte) (uint64, error)
	PacketStoreQoS0(id []byte, packets *PersistedPacket) error
	PacketStoreQoS12(id []byte, packets *PersistedPacket) error
	PacketsForEachQoS0(id []byte, ctx interface{}, loader PacketLoader) error
	PacketsForEachQoS12(id []byte, ctx interface{}, loader PacketLoader) error
	PacketsForEachUnAck(id []byte, ctx interface{}, loader PacketLoader) error
	PacketsStore(id []byte, packets PersistedPackets) error
	PacketsDelete(id []byte) error
}

// Subscriptions session subscriptions interface
type Subscriptions interface {
	SubscriptionsStore([]byte, []byte) error
	SubscriptionsDelete([]byte) error
}

// Expiry session expiration interface
type Expiry interface {
	ExpiryStore(id []byte, delays *SessionDelays) error
	ExpiryDelete(id []byte) error
}

// State session state interface
type State interface {
	StateStore(id []byte, state *SessionState) error
	StateDelete(id []byte) error
}

// Retained provider for load/store retained messages
type Retained interface {
	// Store persist retained message
	// it wipes previously set values
	Store([]*PersistedPacket) error
	// Load load retained messages
	Load() ([]*PersistedPacket, error)
	// Wipe retained storage
	Wipe() error
}

// Sessions interface allows operating with sessions inside backend
type Sessions interface {
	Packets
	Subscriptions
	State
	Expiry
	Create(id []byte, state *SessionBase) error
	Count() uint64
	LoadForEach(loader SessionLoader, ctx interface{}) error
	// Exists check session if presented in storage
	Exists(id []byte) bool
	Delete(id []byte) error
}

// System persistence state of the system configuration
type System interface {
	GetInfo() (*SystemState, error)
	// SetInfo(*SystemState) error
}

// IFace interface implemented by different backends
type IFace interface {
	Sessions() (Sessions, error)
	Retained() (Retained, error)
	System() (System, error)
	Shutdown() error
}
