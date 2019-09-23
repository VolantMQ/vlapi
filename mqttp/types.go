package mqttp

import (
	"bytes"
	"regexp"
	"unicode/utf8"
)

// ProtocolVersion describes versions implemented by this package
type ProtocolVersion byte

const (
	// ProtocolV31 describes spec MQIsdp
	ProtocolV31 = ProtocolVersion(0x3)
	// ProtocolV311 describes spec v3.1.1
	ProtocolV311 = ProtocolVersion(0x4)
	// ProtocolV50 describes spec v5.0
	ProtocolV50 = ProtocolVersion(0x5)
)

var (
	// TopicFilterRegexp regular expression that all subscriptions must be validated
	TopicFilterRegexp = regexp.MustCompile(`^(([^+#]*|\+)(/([^+#]*|\+))*(/#)?|#)$`)

	// TopicPublishRegexp regular expression that all publish to topic must be validated
	TopicPublishRegexp = regexp.MustCompile(`^[^#+]*$`)

	// SharedTopicRegexp regular expression that all share subscription must be validated
	SharedTopicRegexp = regexp.MustCompile(`^\$share/([^#+/]+)(/)(.+)$`)
)

var dollarPrefix = []byte("$")
var sharePrefix = []byte("$share")
var topicSep = []byte("/")

// SupportedVersions is a map of the version number (0x3 or 0x4) to the version string,
// "MQIsdp" for 0x3, and "MQTT" for 0x4.
var SupportedVersions = map[ProtocolVersion]string{
	ProtocolV31:  "MQIsdp",
	ProtocolV311: "MQTT",
	ProtocolV50:  "MQTT",
}

const (
	// MaxLPString maximum size of length-prefixed string
	MaxLPString = 65535
)

// Topic container encapsulates topic verification
type Topic struct {
	full         []byte
	filter       []byte
	dollarPrefix []byte
	shareName    []byte
	ops          SubscriptionOptions
}

// NewTopic allocate new topic
func NewTopic(topic []byte) (*Topic, error) {
	if len(topic) == 0 {
		return nil, CodeMalformedPacket
	}

	// [MQTT-3.10.3-1]
	if !utf8.Valid(topic) {
		return nil, CodeProtocolError
	}

	t := &Topic{
		full:   topic,
		filter: topic,
	}

	if bytes.HasPrefix(topic, dollarPrefix) {
		if idx := bytes.Index(topic, topicSep); idx == 0 {
			t.dollarPrefix = topic[:idx]

			if bytes.Equal(t.dollarPrefix, sharePrefix) {
				if !SharedTopicRegexp.Match(topic) {
					return nil, CodeProtocolError
				}

				sIdx := bytes.Index(topic[idx:], topicSep)
				t.shareName = topic[idx+1 : sIdx]
				t.filter = topic[sIdx+1:]
			}
		} else {
			return nil, CodeProtocolError
		}
	}

	if !TopicFilterRegexp.Match(t.filter) {
		return nil, CodeProtocolError
	}

	return t, nil
}

// NewSubscribeTopic allocate new subscription topic
func NewSubscribeTopic(topic []byte, ops SubscriptionOptions) (*Topic, error) {
	if !ops.QoS().IsValid() {
		return nil, CodeProtocolError
	}

	t, err := NewTopic(topic)
	if err != nil {
		return nil, err
	}

	t.ops = ops

	return t, nil
}

// Full return raw topic
func (t *Topic) Full() string {
	return string(t.full)
}

// Topic return filter and subscription options
func (t *Topic) Topic() (string, SubscriptionOptions) {
	return string(t.filter), t.ops
}

// Filter return topic filter
func (t *Topic) Filter() string {
	return string(t.filter)
}

// Ops return subscription options
func (t *Topic) Ops() SubscriptionOptions {
	return t.ops
}

// DollarPrefix return $ prefix
// if not exist returns empty string
func (t *Topic) DollarPrefix() string {
	return string(t.dollarPrefix)
}

// ShareName return share name
// if not exists returns empty string
func (t *Topic) ShareName() string {
	return string(t.shareName)
}
