package main

import (
	"fmt"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/container"
)

type StartOptions struct {
	*CommonOptions

	// names of the container
	names []string
}

func cli_start(ctx *cli.Context, conf *config.Config) {

	opts := &StartOptions{
		CommonOptions: GetCommonOptions(ctx),
		names:         ctx.Args(),
	}

	for _, name := range opts.names {
		start_container(ctx, conf, name)
	}

}

func start_container(ctx *cli.Context, conf *config.Config, name string) error {
	wait := GetWaitOptions(ctx)

	containerpath := filepath.Join(conf.StoragePath, name)
	logfile := containerpath + ".log"

	c, err := container.NewContainer(containerpath)
	if err != nil {
		log.Error("getting container", name, err)
		return err
	}

	if c.C.Defined() == false {
		log.Error("container not defined", name)
		return fmt.Errorf("container not defined")
	}

	c.LoadMetadata()
	if err != nil {
		log.Error("load container meta", name, err)
		return err
	}

	c.Build.Common()
	c.Build.Logging(logfile, container.LogTrace)

	runtimerootfs := filepath.Join(conf.StoragePath, c.Meta.Runtime, "rootfs")
	c.Build.RootFilesystem(runtimerootfs, c.Path("/rootfs"))

	err = c.Prepare(conf, c.Meta)
	if err != nil {
		log.Errorf("Prepare: %s", err)
		return err
	}

	err = c.Start()
	if err != nil {
		log.Error("container start", name, err)
		return err
	}

	err = c.WaitStart(wait)
	if err != nil {
		log.Error("container wait start", name, err)
		return err
	}
	return err
}
