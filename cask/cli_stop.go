package main

import (
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"gopkg.in/lxc/go-lxc.v2"
)

type StopOptions struct {
	*CommonOptions

	// name of the container
	name string
}

func cli_stop(c *cli.Context, conf *config.Config) {

	opts := &StartOptions{
		CommonOptions: GetCommonOptions(c),
		name:          c.String("name"),
	}

	if opts.name == "" {
		log.Error("container name required")
		return
	}

	container, err := lxc.NewContainer(opts.name, conf.StoragePath)
	if err != nil {
		log.Error("getting container", opts.name, err)
		return
	}

	log.Info(opts.name, "is", container.State())
	err = container.Stop()
	if err != nil {
		log.Error("stopping container", err)
		return
	}
}