package vlsubscriber

import (
	"github.com/VolantMQ/vlapi/mqttp"
)

// SubscriptionParams parameters of the subscription
type SubscriptionParams struct {
	// Subscription id
	// V5.0 ONLY
	ID uint32

	// Ops requested subscription options
	Ops mqttp.SubscriptionOptions

	// Granted QoS granted by topics manager
	Granted mqttp.QosType
}

// Subscriptions contains active subscriptions with respective subscription parameters
type Subscriptions map[string]*SubscriptionParams

// IFace passed to present network connection
type IFace interface {
	Subscriptions() Subscriptions
	Subscribe(string, *SubscriptionParams) (mqttp.QosType, []*mqttp.Publish, error)
	UnSubscribe(string) error
	HasSubscriptions() bool
	Online(c Publisher)
	Offline(bool)
	Hash() uintptr
}

// Publisher implemented by subscribing object.
// By calling Publish function subscriber delivers message
type Publisher func(string, *mqttp.Publish)
