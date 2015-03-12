package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"gopkg.in/lxc/go-lxc.v2"
	"path/filepath"
	"strings"
)

type ListOptions struct {
	*CommonOptions

	// runtime name to build image in, ie "ubuntu12"
	runtime string

	// name of the container
	name string
}

func list(c *cli.Context) {

	opts := &ListOptions{
		CommonOptions: GetCommonOptions(c),
		name:          c.String("name"),
		runtime:       c.String("runtime"),
	}

	runtimepath := filepath.Join(opts.lxcpath, opts.runtime)

	log.Debug("runtime", opts.runtime)
	log.Debug("runtimepath", runtimepath)

	FMT := "%-20s %-10s %-20s\n"
	fmt.Printf(FMT, "NAME", "STATE", "IPV4")
	fmt.Printf("%s\n", strings.Repeat("-", 41))

	containers := lxc.Containers(opts.lxcpath)
	for _, container := range containers {
		ipv4addrs, _ := container.IPv4Addresses()
		var ip string
		if len(ipv4addrs) > 0 {
			ip = ipv4addrs[0]
		}
		fmt.Printf(FMT, container.Name(), container.State(), ip)
	}

}
