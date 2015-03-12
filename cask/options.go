package main

import (
	"github.com/codegangsta/cli"
)

type CommonOptions struct {
	lxcpath string
	verbose bool
}

func GetCommonOptions(c *cli.Context) *CommonOptions {
	opts := &CommonOptions{}
	opts.lxcpath = c.GlobalString("lxcpath")
	opts.verbose = c.GlobalBool("verbose")
	return opts
}
