package broker

import (
	"errors"
	"fmt"
)

var (
	ErrBrokerNotConnected = errors.New("broker not connected")
	ErrCanNotConnect      = func(msg string, err error) error {
		return fmt.Errorf("%s, err: %w", msg, err)
	}
	ErrCanNotSubscribe = func(subject string, err error) error {
		return fmt.Errorf("can not subscribe to topic %s, err: %w", subject, err)
	}
)
