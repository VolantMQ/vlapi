package vlauth

// AccessType acl type
type AccessType int

// Status auth
type Status int

// Error auth provider errors
type Error int

// nolint: golint
const (
	AccessRead  AccessType = 1
	AccessWrite AccessType = 2
)

// nolint: golint
const (
	StatusAllow Status = iota
	StatusDeny
)

// nolint: golint
const (
	ErrInvalidArgs Error = iota
	ErrUnknownProvider
	ErrAlreadyExists
	ErrNotFound
	ErrNotOpen
	ErrInternal
)

var errorsDesc = map[Error]string{
	ErrInvalidArgs:     "auth: invalid arguments",
	ErrUnknownProvider: "auth: unknown provider",
	ErrAlreadyExists:   "auth: already exists",
	ErrNotFound:        "auth: not found",
	ErrNotOpen:         "auth: not open",
	ErrInternal:        "auth: internal error",
}

var statusDesc = map[Status]string{
	StatusAllow: "auth: access granted",
	StatusDeny:  "auth: access denied",
}

// Permissions check session permissions
type Permissions interface {
	// ACL check access type for client id with username
	ACL(clientID, username, topic string, accessType AccessType) error
}

// IFace interface to auth backends
type IFace interface {
	Permissions
	// Password try authenticate with username and password
	Password(clientID, user, password string) error
	// Shutdown provider
	Shutdown() error
}

// Type return string representation of the type
func (t AccessType) Type() string {
	switch t {
	case AccessRead:
		return "read"
	case AccessWrite:
		return "write"
	}

	return ""
}

func (e Error) Error() string {
	if s, ok := errorsDesc[e]; ok {
		return s
	}

	return "auth: unknown error"
}

func (e Status) Error() string {
	if s, ok := statusDesc[e]; ok {
		return s
	}

	return "auth status: unknown status"
}
