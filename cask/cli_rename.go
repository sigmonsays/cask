package main

import (
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"gopkg.in/lxc/go-lxc.v2"
)

type RenameOptions struct {
	*CommonOptions

	// force it
	force bool

	// name of the container
	name string

	// new name of the container
	newname string
}

func cli_rename(c *cli.Context, conf *config.Config) {

	opts := &RenameOptions{
		CommonOptions: GetCommonOptions(c),
		force:         c.Bool("force"),
		name:          c.Args().First(),
		newname:       c.Args().Get(1),
	}

	if opts.name == "" {
		log.Error("container name required")
		return
	}
	if opts.newname == "" {
		log.Error("new container name required")
		return
	}

	container, err := lxc.NewContainer(opts.name, conf.StoragePath)
	if err != nil {
		log.Error("getting container", opts.name, err)
		return
	}

	newcontainer, err := lxc.NewContainer(opts.newname, conf.StoragePath)
	if newcontainer.Defined() {
		log.Error("new container already exists", opts.newname)
		return
	}

	log.Info(opts.name, "is", container.State())

	if container.State() == lxc.RUNNING {
		if opts.force == false {
			log.Info("container is running.")
			return
		}
		err = container.Stop()
		if err != nil {
			log.Error("stopping container", opts.name, err)
			return
		}
	}

	err = container.Rename(opts.newname)
	if err != nil {
		log.Errorf("rename container %s -> %s: %s", opts.name, opts.newname, err)
		return
	}

}
