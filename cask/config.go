package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"gopkg.in/lxc/go-lxc.v2"
	"path/filepath"
)

type ConfigOptions struct {
	*CommonOptions

	// runtime name to build image in, ie "ubuntu12"
	runtime string

	// name of the container
	name string
}

func config(c *cli.Context) {

	opts := &ConfigOptions{
		CommonOptions: GetCommonOptions(c),
		runtime:       c.String("runtime"),
		name:          c.String("name"),
	}

	runtimepath := filepath.Join(opts.lxcpath, opts.runtime)

	fmt.Println("runtime", opts.runtime)
	fmt.Println("runtimepath", runtimepath)

	runtime, err := lxc.NewContainer(opts.runtime, opts.lxcpath)
	if err != nil {
		fmt.Println("ERROR getting runtime container", err)
		return
	}

	fmt.Println("-- runtime configuration --")

	network := runtime.ConfigItem("lxc.network")
	fmt.Println("network", network)

	keys := runtime.ConfigKeys()
	for _, key := range keys {
		values := runtime.ConfigItem(key)
		fmt.Printf("#%s %#v\n", key, values)

		for _, value := range values {
			if value == "" {
				continue
			}
			fmt.Printf("%s = %s\n", key, value)
		}
	}

}
