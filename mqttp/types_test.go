package mqttp

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestUTFInvalid(t *testing.T) {
	for i := byte(0); i <= 0x1F; i++ {
		buf := []byte{i}
		require.Equal(t, false, IsValidUTF(buf))
	}

	for i := byte(0x7F); i <= 0x9F; i++ {
		buf := []byte{i}
		require.Equal(t, false, IsValidUTF(buf))
	}
}

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
	require.NoError(t, err)
	require.NotNil(t, topic)

}
func TestTopicNewValid1(t *testing.T) {
	topic, err := NewTopic([]byte("/"))
	require.NoError(t, err)
	require.NotNil(t, topic)
	require.Equal(t, "/", topic.Filter())
	require.Equal(t, "", topic.DollarPrefix())
	require.Equal(t, "", topic.ShareName())
}
