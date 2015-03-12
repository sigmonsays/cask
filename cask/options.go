package main

import (
	"github.com/codegangsta/cli"
	"time"
)

type CommonOptions struct {
	lxcpath string
	verbose bool

	// time to wait for specific events
	waitTimeout, waitNetworkTimeout time.Duration

	// what states to wait for
	waitMask int
}

const (
	WaitMaskStart = iota
	WaitMaskNetwork
)

func GetCommonOptions(c *cli.Context) *CommonOptions {
	opts := &CommonOptions{}
	opts.lxcpath = c.GlobalString("lxcpath")
	opts.verbose = c.GlobalBool("verbose")
	opts.waitMask = c.GlobalInt("wait")
	opts.waitTimeout = c.GlobalDuration("wait-timeout")
	opts.waitNetworkTimeout = c.GlobalDuration("net-timeout")
	log.Tracef("common options %+v", opts)
	return opts
}
