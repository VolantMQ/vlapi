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
	// TopicSubscribeRegexp regular expression that all subscriptions must be validated
	TopicFilterRegexp = regexp.MustCompile(`^(([^+#]*|\+)(/([^+#]*|\+))*(/#)?|#)$`)

	// TopicPublishRegexp regular expression that all publish to topic must be validated
	TopicPublishRegexp = regexp.MustCompile(`^[^#+]*$`)

	// SharedTopicRegexp regular expression that all share subscription must be validated
	SharedTopicRegexp = regexp.MustCompile(`^\$share\/([^#,+,/]+)(\/)(.+)$`)
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

type Topic struct {
	full         []byte
	filter       []byte
	dollarPrefix []byte
	shareName    []byte
	ops          SubscriptionOptions
}

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

			if bytes.Compare(t.dollarPrefix, sharePrefix) == 0 {
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

func NewSubscribeTopic(topic []byte, ops SubscriptionOptions) (*Topic, error) {
	if !ops.QoS().IsValid() {
		return nil, CodeProtocolError
	}

	t, err := NewTopic(topic)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (t *Topic) Full() string {
	return string(t.full)
}

func (t *Topic) Topic() (string, SubscriptionOptions) {
	return string(t.filter), t.ops
}

func (t *Topic) Filter() string {
	return string(t.filter)
}

func (t *Topic) Ops() SubscriptionOptions {
	return t.ops
}

func (t *Topic) DollarPrefix() string {
	return string(t.dollarPrefix)
}

func (t *Topic) ShareName() string {
	return string(t.shareName)
}
