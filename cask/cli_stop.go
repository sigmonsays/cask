package main

import (
	"fmt"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/container"
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
	containerpath := filepath.Join(conf.StoragePath, name)

	container, err := container.NewContainer(containerpath)
	if err != nil {
		log.Error("getting container", name, err)
		return err
	}

	if container.C.Defined() == false {
		log.Error("container not defined", name)
		return fmt.Errorf("container not defined")
	}

	log.Info(name, "is", container.C.State())
	err = container.C.Stop()
	if err != nil {
		log.Error("stopping container", err)
		return err
	}
	return nil
}
