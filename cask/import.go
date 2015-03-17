package main

import (
	"github.com/codegangsta/cli"
	"gopkg.in/lxc/go-lxc.v2"
)

type ImportOptions struct {
	*CommonOptions

	// name of the container
	name string

	// bootstrap the image using a template
	bootstrap string
}

func Import(c *cli.Context) {

	opts := &StartOptions{
		CommonOptions: GetCommonOptions(c),
		name:          c.String("name"),
	}

	if opts.name == "" {
		log.Error("container name required")
		return
	}

	container, err := lxc.NewContainer(opts.name, opts.lxcpath)
	if err != nil {
		log.Error("getting container", opts.name, err)
		return
	}

	log.Info(opts.name, "is", container.State())
	log.Info("TODO")
}
