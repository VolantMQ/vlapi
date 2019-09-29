// Copyright (c) 2014 The VolantMQ Authors. All rights reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mqttp

// UnSubscribe An UNSUBSCRIBE Packet is sent by the Client to the Server, to unsubscribe from topics.
type UnSubscribe struct {
	header

	topics []*Topic
}

var _ IFace = (*UnSubscribe)(nil)

func newUnSubscribe() *UnSubscribe {
	return &UnSubscribe{}
}

// NewUnSubscribe creates a new UNSUBSCRIBE packet
func NewUnSubscribe(v ProtocolVersion) *UnSubscribe {
	p := newUnSubscribe()
	p.init(UNSUBSCRIBE, v, p.size, p.encodeMessage, p.decodeMessage)
	return p
}

// ForEachTopic iterate over all sent by the Client
// if fn return error iteration stopped
func (msg *UnSubscribe) ForEachTopic(fn func(*Topic) error) error {
	for _, t := range msg.topics {
		if err := fn(t); err != nil {
			return err
		}
	}

	return nil
}

// AddTopic adds a single topic to the message.
func (msg *UnSubscribe) AddTopic(topic *Topic) error {
	msg.topics = append(msg.topics, topic)

	return nil
}

// SetPacketID sets the ID of the packet.
func (msg *UnSubscribe) SetPacketID(v IDType) {
	msg.setPacketID(v)
}

// decode reads from the io.Reader parameter until a full message is decoded, or
// when io.Reader returns EOF or error. The first return value is the number of
// bytes read from io.Reader. The second is error if decode encounters any problems.
func (msg *UnSubscribe) decodeMessage(from []byte) (int, error) {
	offset := msg.decodePacketID(from)

	remLen := int(msg.remLen) - offset

	// V5.0
	// [MQTT-3.10.2.1] UNSUBSCRIBE Properties
	if msg.version >= ProtocolV50 {
		n, err := msg.properties.decode(msg.Type(), from[offset:])
		offset += n
		if err != nil {
			return offset, err
		}
		remLen -= n
	}

	for remLen > 0 {
		t, n, err := ReadLPBytes(from[offset:])
		offset += n
		if err != nil {
			return offset, err
		}

		var topic *Topic
		if topic, err = NewTopic(t); err != nil {
			return offset, err
		}

		msg.topics = append(msg.topics, topic)

		remLen = remLen - n - 1
	}

	// [MQTT-3.10.3-2]
	if len(msg.topics) == 0 {
		rejectReason := CodeProtocolError
		if msg.version <= ProtocolV50 {
			rejectReason = CodeRefusedServerUnavailable
		}
		return 0, rejectReason
	}

	return offset, nil
}

func (msg *UnSubscribe) encodeMessage(dst []byte) (int, error) {
	// [MQTT-2.3.1]
	if len(msg.packetID) == 0 {
		return 0, ErrPackedIDZero
	}

	if len(msg.topics) == 0 {
		return 0, ErrInvalidLength
	}

	var err error
	var n int

	total := msg.encodePacketID(dst)

	for _, t := range msg.topics {
		n, err = WriteLPBytes(dst[total:], t.full)
		total += n
		if err != nil {
			return total, err
		}
	}

	return total, err
}

// Encode returns an io.Reader in which the encoded bytes can be read. The second
// return value is the number of bytes encoded, so the caller knows how many bytes
// there will be. If Encode returns an error, then the first two return values
// should be considered invalid.
// Any changes to the message after Encode() is called will invalidate the io.Reader.

func (msg *UnSubscribe) size() int {
	// packet ID
	total := 2

	for _, t := range msg.topics {
		total += 2 + len(t.full)
	}

	return total
}
