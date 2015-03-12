package main

import (
	"github.com/codegangsta/cli"
	"gopkg.in/lxc/go-lxc.v2"
)

type StopOptions struct {
	*CommonOptions

	// name of the container
	name string
}

func stop(c *cli.Context) {

	opts := &StartOptions{
		CommonOptions: GetCommonOptions(c),
		name:          c.String("name"),
	}

	container, err := lxc.NewContainer(opts.name, opts.lxcpath)
	if err != nil {
		log.Error("ERROR getting container", opts.name, err)
		return
	}

	log.Info(opts.name, "is", container.State())
	err = container.Stop()
	if err != nil {
		log.Error("ERROR stopping container", err)
		return
	}
}
