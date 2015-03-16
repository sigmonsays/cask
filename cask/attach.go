package main

import (
	"github.com/codegangsta/cli"
	"gopkg.in/lxc/go-lxc.v2"
)

type AttachOptions struct {
	*CommonOptions

	// name of the container
	name string
}

func attach(c *cli.Context) {

	opts := &AttachOptions{
		CommonOptions: GetCommonOptions(c),
		name:          c.String("name"),
	}

	container, err := lxc.NewContainer(opts.name, opts.lxcpath)
	if err != nil {
		log.Error("getting container", opts.name, err)
		return
	}

	log.Info(opts.name, "is", container.State())

	options := lxc.DefaultAttachOptions

	err = container.AttachShell(options)
	if err != nil {
		log.Error("attaching container", opts.name, err)
		return
	}
}
