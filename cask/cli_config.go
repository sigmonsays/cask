package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"gopkg.in/lxc/go-lxc.v2"
)

type ConfigOptions struct {
	*CommonOptions

	// runtime name to build image in, ie "ubuntu12"
	runtime string

	// name of the container
	name string
}

func cli_config(c *cli.Context, conf *config.Config) {

	opts := &ConfigOptions{
		CommonOptions: GetCommonOptions(c),
		runtime:       c.String("runtime"),
		name:          c.String("name"),
	}

	container, err := lxc.NewContainer(opts.name, conf.StoragePath)
	if err != nil {
		log.Error("getting container", err)
		return
	}
	keys := container.ConfigKeys()
	for _, key := range keys {
		values := container.ConfigItem(key)

		for _, value := range values {
			if value == "" {
				continue
			}
			fmt.Printf("[container] %s = %s\n", key, value)
		}
	}

	if opts.verbose {
		runtime, err := lxc.NewContainer(opts.runtime, conf.StoragePath)
		if err != nil {
			log.Error("getting runtime container", err)
			return
		}
		keys := runtime.ConfigKeys()
		for _, key := range keys {
			values := runtime.ConfigItem(key)

			for _, value := range values {
				if value == "" {
					continue
				}
				fmt.Printf("[runtime] %s = %s\n", key, value)
			}
		}
	}

}
