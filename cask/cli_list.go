package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"gopkg.in/lxc/go-lxc.v2"
	"path/filepath"
	"strings"
)

type ListOptions struct {
	*CommonOptions

	// runtime name to build image in, ie "ubuntu12"
	runtime string

	// show all containers (ie, include stopped)
	all bool
}

func cli_list(c *cli.Context, conf *config.Config) {

	opts := &ListOptions{
		CommonOptions: GetCommonOptions(c),
		runtime:       c.String("runtime"),
		all:           c.Bool("all"),
	}

	runtimepath := filepath.Join(conf.StoragePath, opts.runtime)

	log.Debug("runtime", opts.runtime)
	log.Debug("runtimepath", runtimepath)

	FMT := "%-20s %-10s %-20s\n"
	fmt.Printf(FMT, "NAME", "STATE", "IPV4")
	fmt.Printf("%s\n", strings.Repeat("-", 41))

	containers := lxc.Containers(conf.StoragePath)
	for _, container := range containers {
		if opts.all == false && container.Running() == false {
			continue
		}
		ipv4addrs, _ := container.IPv4Addresses()
		var ip string
		if len(ipv4addrs) > 0 {
			ip = ipv4addrs[0]
		}
		fmt.Printf(FMT, container.Name(), container.State(), ip)
	}

}
