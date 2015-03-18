package main

import (
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/container"
	"path/filepath"
)

type StartOptions struct {
	*CommonOptions

	// name of the container
	name string
}

func cli_start(ctx *cli.Context, conf *config.Config) {

	opts := &StartOptions{
		CommonOptions: GetCommonOptions(ctx),
		name:          ctx.String("name"),
	}

	wait := GetWaitOptions(ctx)

	if opts.name == "" {
		log.Error("container name required")
		return
	}

	containerpath := filepath.Join(conf.StoragePath, opts.name)

	c, err := container.NewContainer(containerpath)
	if err != nil {
		log.Error("getting container", opts.name, err)
		return
	}

	c.LoadMetadata()
	if err != nil {
		log.Error("load container meta", opts.name, err)
		return
	}

	err = c.Start()
	if err != nil {
		log.Error("container start", opts.name, err)
		return
	}

	err = c.WaitStart(wait)
	if err != nil {
		log.Error("container wait start", opts.name, err)
		return
	}

}
