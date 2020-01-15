package vlmonitoring

import (
	"sync/atomic"
)

type base struct {
	uint64
}

type Counter struct {
	base
}

type FlowCounter struct {
	Sent Counter
	Recv Counter
}

type State struct {
	base
}

type Max struct {
	State
	Max uint64
}

type Bytes struct {
	FlowCounter
}

type Packets struct {
	Total      FlowCounter
	Connect    Counter
	ConnAck    Counter
	Publish    FlowCounter
	Puback     FlowCounter
	Pubrec     FlowCounter
	Pubrel     FlowCounter
	Pubcomp    FlowCounter
	Sub        Counter
	SubAck     Counter
	UnSub      Counter
	UnSubAck   Counter
	Disconnect FlowCounter
	PingReq    Counter
	PingResp   Counter
	Auth       FlowCounter
	Unknown    Counter
	Rejected   Counter
	UnAckSent  Max
	UnAckRecv  Max
	Retained   Max
	Stored     Max
}

type Clients struct {
	Connected Max
	Persisted Max
	Total     Max
	Expired   Counter
	Rejected  Counter
}

type Subscriptions struct {
	Total Max
}

type Stats struct {
	Bytes   Bytes
	Packets Packets
	Clients Clients
	Subs    Subscriptions
}

func (c *Counter) AddU64(n uint64) uint64 {
	return atomic.AddUint64(&c.uint64, n)
}

func (c *Counter) Diff(prev *Counter) Counter {
	return Counter{
		base: base{
			LoadStore(&c.base.uint64, &prev.base.uint64),
		},
	}
}

func (c *FlowCounter) SentAddU64(n uint64) uint64 {
	return c.Sent.AddU64(n)
}

func (c *FlowCounter) RecvAddU64(n uint64) uint64 {
	return c.Recv.AddU64(n)
}

func (c *FlowCounter) Diff(prev *FlowCounter) FlowCounter {
	return FlowCounter{
		Sent: Counter{
			base{LoadStore(&c.Sent.base.uint64, &prev.Sent.base.uint64)},
		},
		Recv: Counter{
			base{LoadStore(&c.Recv.base.uint64, &prev.Recv.base.uint64)},
		},
	}
}

func (c *base) Load() uint64 {
	return atomic.LoadUint64(&c.uint64)
}

func (c *base) Get() uint64 {
	return c.uint64
}

func (c *Max) Load() Max {
	return Max{
		State: State{
			base: base{
				uint64: atomic.LoadUint64(&c.uint64),
			},
		},
		Max: atomic.LoadUint64(&c.Max),
	}
}

func LoadStore(src *uint64, dst *uint64) uint64 {
	val := atomic.LoadUint64(src)
	diff := val - *dst
	if diff != 0 {
		*dst = val
	}

	return diff
}

func (c *Max) AddU64(n uint64) {
	val := atomic.AddUint64(&c.base.uint64, n)
	if atomic.LoadUint64(&c.Max) < val {
		atomic.StoreUint64(&c.Max, val)
	}
}

func (c *Max) SubU64(n uint64) {
	atomic.AddUint64(&c.base.uint64, ^(n - 1))
}
