package stan

import (
	"time"
)

const (
	DefaultAckWait    = time.Minute
	DefaultMaxDeliver = 30
)

type AckOption func(cfg *ackConfig)

type ackConfig struct {
	ackWait    time.Duration
	maxDeliver int
}

func defaultAckConfig() ackConfig {
	return ackConfig{
		ackWait:    DefaultAckWait,
		maxDeliver: DefaultMaxDeliver,
	}
}
