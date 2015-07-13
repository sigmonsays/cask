package main

import (
	"os"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"gopkg.in/lxc/go-lxc.v2"
)

type ExecOptions struct {
	*CommonOptions

	// name of the container
	name string

	// command line to run
	cmdline []string

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

func cli_exec(c *cli.Context, conf *config.Config) {

	opts := &ExecOptions{
		CommonOptions: GetCommonOptions(c),
		name:          c.Args().First(),
		cmdline:       c.Args()[1:],
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

	if len(opts.cmdline) == 0 {
		log.Error("command line required")
		return
	}

	if opts.cmdline[0] == "--" {
		opts.cmdline = opts.cmdline[1:]
	}

	container, err := lxc.NewContainer(opts.name, conf.StoragePath)
	if err != nil {
		log.Error("getting container", opts.name, err)
		return
	}

	log.Info(opts.name, "is", container.State())

	if container.Running() == false {
		log.Error("container must be running to use exec, it is", container.State())
		return
	}

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

	log.Debugf("cmdline %s", opts.cmdline)
	code, err := container.RunCommandStatus(opts.cmdline, options)
	if err != nil {
		log.Error("attaching container", opts.name, err)
	}

	os.Exit(code)
}
