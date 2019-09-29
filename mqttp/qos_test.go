package mqttp

import "testing"

func TestQosCodes(t *testing.T) {
	if QoS0 != 0 || QoS1 != 1 || QoS2 != 2 {
		t.Errorf("QOS codes invalid")
	}
}
