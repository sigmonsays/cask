package main

import (
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"gopkg.in/lxc/go-lxc.v2"
)

type DestroyOptions struct {
	*CommonOptions

	names []string
}

func cli_destroy(c *cli.Context, conf *config.Config) {

	opts := &DestroyOptions{
		CommonOptions: GetCommonOptions(c),
		names:         c.Args(),
	}

	for _, name := range opts.names {
		destroy_container(c, conf, name)
	}

}

func destroy_container(c *cli.Context, conf *config.Config, name string) error {
	container, err := lxc.NewContainer(name, conf.StoragePath)
	if err != nil {
		log.Error("getting container", name, err)
		return err
	}

	log.Info(name, "is", container.State())

	if container.State() == lxc.RUNNING {
		container.Stop()
	}

	err = container.Destroy()
	if err != nil {
		log.Error("stopping container", name, err)
		return err
	}
	return nil
}
