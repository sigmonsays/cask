package main

import (
	"os"
	"path/filepath"

	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/container"
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

func destroy_container(ctx *cli.Context, conf *config.Config, name string) error {
	containerpath := filepath.Join(conf.StoragePath, name)

	c, err := container.NewContainer(containerpath)
	if err != nil {
		log.Error("getting container", name, err)
		return err
	}

	log.Info(name, "is", c.C.State())

	if c.C.State() == lxc.RUNNING {
		c.C.Stop()
	}

	err = os.RemoveAll(containerpath)
	if err != nil {
		log.Warnf("RemoveAll %s: %s", containerpath, err)
	}

	/*
		// This destroys referenced rootfs so we shouldn't use it
		err = container.Destroy()
		if err != nil {
			log.Error("stopping container", name, err)
			return err
		}
	*/
	return nil
}
