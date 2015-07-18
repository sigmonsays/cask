package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/container"
	"gopkg.in/lxc/go-lxc.v2"
)

type CheckpointOptions struct {
	*CommonOptions

	stop bool

	// name of the container
	name string

	directory string
}

func cli_checkpoint(ctx *cli.Context, conf *config.Config) {

	opts := &CheckpointOptions{
		CommonOptions: GetCommonOptions(ctx),
		name:          ctx.Args().First(),
		stop:          ctx.Bool("stop"),
		directory:     ctx.String("directory"),
	}

	checkpoint_container(ctx, conf, opts, opts.name)
}

func checkpoint_container(ctx *cli.Context, conf *config.Config, opts *CheckpointOptions, name string) error {

	containerpath := filepath.Join(conf.StoragePath, name)

	c, err := container.NewContainer(containerpath)
	if err != nil {
		log.Error("getting container", name, err)
		return err
	}

	copts := lxc.CheckpointOptions{
		Stop:      opts.stop,
		Verbose:   opts.verbose,
		Directory: opts.directory,
	}

	st, err := os.Stat(opts.directory)
	if err != nil {
		err = os.MkdirAll(opts.directory, 0755)
		if err != nil {
			log.Errorf("mkdir %s: %s", opts.directory, err)
			return err
		}
	}

	if st.IsDir() == false {
		log.Errorf("%s is not a directory.", opts.directory)
		return fmt.Errorf("not a directory")
	}

	err = c.C.Checkpoint(copts)
	if err != nil {
		log.Error("checkpoint container", name, err)
		return err
	}

	return nil
}
