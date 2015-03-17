package main

import (
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"gopkg.in/lxc/go-lxc.v2"
	"strings"
	"syscall"
)

type AttachOptions struct {
	*CommonOptions

	// name of the container
	name string

	// attach options
	namespaces string
	arch       string
	cwd        string
	uid, gid   int
	clear_env  bool
	env        []string
	keep_env   []string
	// stdin, stdout, stderr

}

var namespace_names = map[string]int{
	"ipc":  syscall.CLONE_NEWNS,
	"net":  syscall.CLONE_NEWNET,
	"ns":   syscall.CLONE_NEWNS,
	"pid":  syscall.CLONE_NEWPID,
	"user": syscall.CLONE_NEWUSER,
	"uts":  syscall.CLONE_NEWUTS,
}

func cli_attach(c *cli.Context, conf *config.Config) {

	opts := &AttachOptions{
		CommonOptions: GetCommonOptions(c),
		name:          c.String("name"),
		namespaces:    c.String("namespaces"),
		cwd:           c.String("cwd"),
		uid:           c.Int("uid"),
		gid:           c.Int("gid"),
		clear_env:     c.Bool("clear_env"),
		keep_env:      make([]string, 0),
	}

	if opts.name == "" {
		log.Error("container name required")
		return
	}

	container, err := lxc.NewContainer(opts.name, conf.StoragePath)
	if err != nil {
		log.Error("getting container", opts.name, err)
		return
	}

	log.Info(opts.name, "is", container.State())

	options := lxc.DefaultAttachOptions
	if len(opts.namespaces) > 0 {
		namespaces := strings.Fields(opts.namespaces)
		for _, namespace := range namespaces {
			n, ok := namespace_names[namespace]
			if ok == false {
				log.Error("invalid namespace", namespace)
				return
			}
			options.Namespaces |= n
		}
	}
	options.Cwd = opts.cwd
	options.UID = opts.uid
	options.GID = opts.gid
	options.ClearEnv = opts.clear_env
	options.EnvToKeep = opts.keep_env

	err = container.AttachShell(options)
	if err != nil {
		log.Error("attaching container", opts.name, err)
		return
	}
}
