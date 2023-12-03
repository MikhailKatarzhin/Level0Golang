package stan

import (
	"errors"
)

var (
	ErrExistsServerName   = errors.New("server name does not match")
	ErrCanNotToConnect    = errors.New("can not to connect to stan")
	ErrCanNotSubscribe    = errors.New("can not to subscribe to stan")
	ErrCanNotToDisconnect = errors.New("can not to disconnect from stan")
)
