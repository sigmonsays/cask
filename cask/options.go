package main

import (
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/container"
	"time"
)

type CommonOptions struct {
	verbose bool

	// time to wait for specific events
	waitTimeout, waitNetworkTimeout time.Duration

	// what states to wait for
	waitMask int
}

func GetCommonOptions(c *cli.Context) *CommonOptions {
	opts := &CommonOptions{}
	opts.verbose = c.GlobalBool("verbose")
	opts.waitMask = c.GlobalInt("wait")
	opts.waitTimeout = c.GlobalDuration("wait-timeout")
	opts.waitNetworkTimeout = c.GlobalDuration("net-timeout")
	log.Tracef("common options %+v", opts)
	return opts
}

func GetWaitOptions(c *cli.Context) *container.WaitOptions {
	opts := &container.WaitOptions{
		WaitMask:           c.GlobalInt("wait"),
		WaitTimeout:        c.GlobalDuration("wait-timeout"),
		WaitNetworkTimeout: c.GlobalDuration("net-timeout"),
	}
	return opts
}
