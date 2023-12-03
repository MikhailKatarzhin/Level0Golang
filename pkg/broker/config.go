package broker

import (
	"time"
)

const (
	//-1 means infinity retrying
	DefaultMaxReconnects     = -1
	DefaultQueueLength       = 128
	DefaultReconnectsTimeout = time.Second
)

type NATSConfig struct {
	Addr                 string
	User                 string
	Password             string
	MaxReconnects        int
	ReconnectsTimeout    time.Duration
	SubscriptionQueueLen int
}
