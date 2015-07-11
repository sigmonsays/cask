package main

import (
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"gopkg.in/lxc/go-lxc.v2"
)

type StopOptions struct {
	*CommonOptions

	// name of the container
	names []string
}

func cli_stop(ctx *cli.Context, conf *config.Config) {

	opts := &StopOptions{
		CommonOptions: GetCommonOptions(ctx),
		names:         ctx.Args(),
	}

	if len(opts.names) == 0 {
		log.Error("one container name required")
		return
	}

	for _, name := range opts.names {
		stop_container(ctx, conf, name)
	}
}

func stop_container(ctx *cli.Context, conf *config.Config, name string) error {
	container, err := lxc.NewContainer(name, conf.StoragePath)
	if err != nil {
		log.Error("getting container", name, err)
		return err
	}

	log.Info(name, "is", container.State())
	err = container.Stop()
	if err != nil {
		log.Error("stopping container", err)
		return err
	}
	return nil
}
