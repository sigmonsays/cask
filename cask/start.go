package main

import (
	"github.com/codegangsta/cli"
	"gopkg.in/lxc/go-lxc.v2"
	"path/filepath"
)

type StartOptions struct {
	*CommonOptions

	// runtime name to build image in, ie "ubuntu12"
	runtime string

	// name of the container
	name string
}

func start(c *cli.Context) {

	opts := &StartOptions{
		CommonOptions: GetCommonOptions(c),
		runtime:       c.String("runtime"),
		name:          c.String("name"),
	}

	runtimepath := filepath.Join(opts.lxcpath, opts.runtime)

	log.Debug("runtime", opts.runtime)
	log.Debug("runtimepath", runtimepath)

	container, err := lxc.NewContainer(opts.name, opts.lxcpath)
	if err != nil {
		log.Error("ERROR getting runtime container", err)
		return
	}

	log.Info(opts.name, "is", container.State())
}
