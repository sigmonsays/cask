package main

import (
	"flag"
	"fmt"
	"gopkg.in/lxc/go-lxc.v2"
	"os"
	"path/filepath"
	"strings"
)

type ListOptions struct {

	// be more verbose in some cases
	verbose bool

	// runtime name to build image in, ie "ubuntu12"
	runtime string

	// lxcpath where lxc config is stored, ie /var/lib/lxc
	lxcpath string

	// name of the container
	name string
}

func list() {

	opts := &ListOptions{}

	f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
	f.BoolVar(&opts.verbose, "verbose", false, "be verbose")
	f.StringVar(&opts.lxcpath, "lxcpath", lxc.DefaultConfigPath(), "Use specified container path")
	f.Parse(os.Args[2:])

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
