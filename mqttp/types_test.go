package mqttp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTopicNewInvalidArgs(t *testing.T) {
	topic, err := NewTopic([]byte{})
	require.Error(t, err)
	require.Error(t, CodeMalformedPacket, err.Error())
	require.Nil(t, topic)
}

func TestTopicNewInvalidWildcard(t *testing.T) {
	topic, err := NewTopic([]byte("##"))
	require.Error(t, err, CodeProtocolError)
	require.Nil(t, topic)

	topic, err = NewTopic([]byte("#+"))
	require.Error(t, err, CodeProtocolError)
	require.Nil(t, topic)

	topic, err = NewTopic([]byte("+#"))
	require.Error(t, err, CodeProtocolError)
	require.Nil(t, topic)

	topic, err = NewTopic([]byte("#/"))
	require.Error(t, err, CodeProtocolError)
	require.Nil(t, topic)

	topic, err = NewTopic([]byte("/#/"))
	require.Error(t, err, CodeProtocolError)
	require.Nil(t, topic)

	topic, err = NewTopic([]byte("++"))
	require.Error(t, err, CodeProtocolError)
	require.Nil(t, topic)
}

func TestTopicNewInvalidShared(t *testing.T) {
	topic, err := NewTopic([]byte("$share"))
	require.Error(t, err, CodeProtocolError)
	require.Nil(t, topic)

	topic, err = NewTopic([]byte("$share//"))
	require.Error(t, err, CodeProtocolError)
	require.Nil(t, topic)

	topic, err = NewTopic([]byte("$share///ads"))
	require.Error(t, err, CodeProtocolError)
	require.Nil(t, topic)

	topic, err = NewTopic([]byte("$share/#/ads"))
	require.Error(t, err, CodeProtocolError)
	require.Nil(t, topic)

	topic, err = NewTopic([]byte("$share/+/ads"))
	require.Error(t, err, CodeProtocolError)
	require.Nil(t, topic)
}

func TestTopicNewValidShared(t *testing.T) {
	topic, err := NewTopic([]byte("$share/a/ads"))
	require.Error(t, err, CodeProtocolError)
	require.Nil(t, topic)

}
func TestTopicNewValid1(t *testing.T) {
	topic, err := NewTopic([]byte("/"))
	require.NoError(t, err)
	require.NotNil(t, topic)
	require.Equal(t, "/", topic.Filter())
	require.Equal(t, "", topic.DollarPrefix())
	require.Equal(t, "", topic.ShareName())
}
