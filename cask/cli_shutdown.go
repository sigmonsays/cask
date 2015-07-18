package main

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/container"
)

type ShutdownOptions struct {
	*CommonOptions

	timeout time.Duration

	// names of the container
	names []string
}

func cli_shutdown(ctx *cli.Context, conf *config.Config) {

	opts := &ShutdownOptions{
		CommonOptions: GetCommonOptions(ctx),
		names:         ctx.Args(),
		timeout:       ctx.Duration("timeout"),
	}

	for _, name := range opts.names {
		shutdown_container(ctx, conf, opts, name)
	}

}

func shutdown_container(ctx *cli.Context, conf *config.Config, opts *ShutdownOptions, name string) error {

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

	err = c.C.Shutdown(opts.timeout)
	if err != nil {
		log.Error("container shutdown", name, err)
		log.Errorf("check log %s for details", logfile)
		return err
	}

	return nil
}
