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
	names []string
}

func cli_config(c *cli.Context, conf *config.Config) {

	opts := &ConfigOptions{
		CommonOptions: GetCommonOptions(c),
		runtime:       c.String("runtime"),
		names:         c.StringSlice("name"),
	}

	for _, name := range opts.names {
		fmt.Println("")
		show_container_config(conf, opts, name)
	}
}

func show_container_config(conf *config.Config, opts *ConfigOptions, name string) {

	container, err := lxc.NewContainer(name, conf.StoragePath)
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
			fmt.Printf("[container.%s] %s = %s\n", name, key, value)
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
				fmt.Printf("[runtime.%s] %s = %s\n", name, key, value)
			}
		}
	}

}
