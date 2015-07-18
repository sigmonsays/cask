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

type RestoreOptions struct {
	*CommonOptions

	// name of the container
	name string

	directory string
}

func cli_restore(ctx *cli.Context, conf *config.Config) {

	opts := &RestoreOptions{
		CommonOptions: GetCommonOptions(ctx),
		name:          ctx.Args().First(),
		directory:     ctx.String("directory"),
	}

	restore_container(ctx, conf, opts, opts.name)
}

func restore_container(ctx *cli.Context, conf *config.Config, opts *RestoreOptions, name string) error {

	containerpath := filepath.Join(conf.StoragePath, name)

	c, err := container.NewContainer(containerpath)
	if err != nil {
		log.Error("getting container", name, err)
		return err
	}

	ropts := lxc.RestoreOptions{
		Verbose:   opts.verbose,
		Directory: opts.directory,
	}

	st, err := os.Stat(opts.directory)
	if err != nil {
		log.Errorf("stat %s: %s", opts.directory, err)
		return err
	}
	if st.IsDir() == false {
		return fmt.Errorf("not a directory: %s", opts.directory)
	}

	err = c.C.Restore(ropts)
	if err != nil {
		log.Error("restore container", name, err)
		return err
	}

	return nil
}
