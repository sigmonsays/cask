package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"

	"github.com/codegangsta/cli"
	"launchpad.net/goyaml"

	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/container"
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
		names:         c.Args(),
	}

	for _, name := range opts.names {
		fmt.Println("")
		show_container_config(conf, opts, name)
	}
}

func show_container_config(conf *config.Config, opts *ConfigOptions, name string) {
	print_yaml := false

	containerpath := filepath.Join(conf.StoragePath, name)

	c, err := container.NewContainer(containerpath)
	if err != nil {
		log.Error("getting container", name, err)
		return
	}

	c.LoadMetadata()
	if err != nil {
		log.Error("load container meta", name, err)
		return
	}

	if print_yaml {
		blob, err := goyaml.Marshal(c.Meta)
		if err != nil {
			log.Error("yaml marshal meta", name, err)
			return
		}
		fmt.Printf("%s", blob)

	}
	// output json
	blob, err := json.MarshalIndent(c.Meta, "", "   ")
	if err != nil {
		log.Error("json marshal meta", name, err)
		return
	}
	fmt.Printf("%s", blob)

}
