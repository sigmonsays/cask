package container

import (
	"time"
)

type WaitOptions struct {

	// time to wait for specific events
	WaitTimeout, WaitNetworkTimeout time.Duration

	// what states to wait for
	WaitMask int
}

const (
	WaitMaskStart = iota
	WaitMaskNetwork
)
