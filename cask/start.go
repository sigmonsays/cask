package main

import (
	"github.com/codegangsta/cli"
	"gopkg.in/lxc/go-lxc.v2"
)

type StartOptions struct {
	*CommonOptions

	// name of the container
	name string
}

func start(c *cli.Context) {

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

	err = container.Start()
	if err != nil {
		log.Error("ERROR starting container", opts.name, err)
		return
	}
}
