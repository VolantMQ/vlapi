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

	Flags struct {
		// UnAck if packet waits acknowledgment from client
		UnAck bool
	}
}

// PersistedPackets array of persisted packets
type PersistedPackets []*PersistedPacket

// SessionDelays formerly known as expiry set timestamp to handle will delay and/or expiration
type SessionDelays struct {
	Since    string
	ExpireIn string
	// WillIn timestamp when will publish must occur
	WillIn string
	// WillData encoded packet
	WillData []byte
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
	Version  string
	NodeName string
}

// PacketLoader interface must be implemented by session
type PacketLoader interface {
	LoadPersistedPacket(*PersistedPacket) error
}

// SessionLoader implemented by session manager to load persisted sessions when server starts
type SessionLoader interface {
	LoadSession(interface{}, []byte, *SessionState) error
}

// Packets interface for connection to handle packets
type Packets interface {
	PacketsForEach([]byte, PacketLoader) error
	PacketsStore([]byte, PersistedPackets) error
	PacketsDelete([]byte) error
}

// Subscriptions session subscriptions interface
type Subscriptions interface {
	SubscriptionsStore([]byte, []byte) error
	SubscriptionsDelete([]byte) error
}

// Expiry session expiration interface
type Expiry interface {
	ExpiryStore([]byte, *SessionDelays) error
	ExpiryDelete([]byte) error
}

// State session state interface
type State interface {
	StateStore([]byte, *SessionState) error
	StateDelete([]byte) error
}

// Retained provider for load/store retained messages
type Retained interface {
	// Store persist retained message
	Store(PersistedPackets) error
	// Load load retained messages
	Load() (PersistedPackets, error)
	// Wipe retained storage
	Wipe() error
}

// Sessions interface allows operating with sessions inside backend
type Sessions interface {
	Packets
	Subscriptions
	State
	Expiry
	Create([]byte, *SessionBase) error
	Count() uint64
	LoadForEach(SessionLoader, interface{}) error
	PacketStore([]byte, *PersistedPacket) error
	Exists([]byte) bool
	Delete([]byte) error
}

// System persistence state of the system configuration
type System interface {
	GetInfo() (*SystemState, error)
	SetInfo(*SystemState) error
}

// Provider interface implemented by different backends
type Provider interface {
	Sessions() (Sessions, error)
	Retained() (Retained, error)
	System() (System, error)
	Shutdown() error
}
