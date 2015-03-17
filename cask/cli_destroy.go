package main

import (
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"gopkg.in/lxc/go-lxc.v2"
)

type DestroyOptions struct {
	*CommonOptions

	// name of the container
	name string
}

func cli_destroy(c *cli.Context, conf *config.Config) {

	opts := &DestroyOptions{
		CommonOptions: GetCommonOptions(c),
		name:          c.String("name"),
	}

	container, err := lxc.NewContainer(opts.name, opts.lxcpath)
	if err != nil {
		log.Error("getting container", opts.name, err)
		return
	}

	log.Info(opts.name, "is", container.State())

	if container.State() == lxc.RUNNING {
		container.Stop()
	}

	err = container.Destroy()
	if err != nil {
		log.Error("stopping container", opts.name, err)
		return
	}
}
