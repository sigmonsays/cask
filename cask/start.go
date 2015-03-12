package main

import (
	"github.com/codegangsta/cli"
	"gopkg.in/lxc/go-lxc.v2"
	"path/filepath"
)

type StartOptions struct {

	// be more verbose in some cases
	verbose bool

	// runtime name to build image in, ie "ubuntu12"
	runtime string

	// lxcpath where lxc config is stored, ie /var/lib/lxc
	lxcpath string

	// name of the container
	name string
}

func start(c *cli.Context) {

	opts := &StartOptions{}

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
