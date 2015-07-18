package main

import (
	"fmt"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"time"

	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/sup"
	gologging "github.com/sigmonsays/go-logging"
	"gopkg.in/lxc/go-lxc.v2"
)

func WarnIf(err error) {
	if err != nil {
		log.Warnf("%s", err)
	}
}

func main() {
	// different behavior based on our invoked name
	name := filepath.Base(os.Args[0])
	if name == "cask" {
		main_cask()
	} else if name == "cask-init" {
		main_init()
	} else {
		fmt.Println("Unknown invoked name %s", name)
		os.Exit(1)
	}
}

func main_init() {
	sup.Main()
}

func main_cask() {

	conf := config.DefaultConfig()

	CheckPrerequisites()

	app := cli.NewApp()
	app.Name = "cask"
	app.Usage = "manage container lifecycle"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "config, c",
			Value: config.DefaultConfigPath(),
		},
		cli.StringFlag{
			Name:  "storage, s",
			Usage: "override storage path",
		},
		cli.BoolFlag{
			Name:  "verbose",
			Usage: "be verbose",
		},
		cli.StringFlag{
			Name:  "level, l",
			Value: "WARN",
			Usage: "change log level",
		},
		cli.IntFlag{
			Name:  "wait",
			Usage: "wait max",
			Value: 2,
		},
		cli.DurationFlag{
			Name:  "wait-timeout, w",
			Value: time.Duration(3) * time.Second,
			Usage: "how long to wait for container to start",
		},
		cli.DurationFlag{
			Name:  "net-timeout",
			Value: time.Duration(15) * time.Second,
			Usage: "how long to wait for container network to start",
		},
	}
	app.Before = func(c *cli.Context) error {
		gologging.SetLogLevel(c.String("level"))
		conf.FromFile(c.String("config"))
		if c.GlobalIsSet("storagepath") {
			conf.StoragePath = c.GlobalString("storagepath")

		} else if conf.StoragePath == "" {
			conf.StoragePath = lxc.DefaultConfigPath()
		}

		return nil
	}

	app.Commands = []cli.Command{
		{
			Name:  "build",
			Usage: "build a image",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "caskpath",
					Value: "cask",
				},
				cli.StringFlag{
					Name:  "runtime",
					Value: "",
				},
				cli.BoolFlag{
					Name: "keep",
				},
			},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_build)
			},
		},
		{
			Name:  "config",
			Usage: "manage container config",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "runtime",
					Value: "",
				},
				cli.BoolFlag{
					Name: "yaml, y",
				},
			},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_config)
			},
		},
		{
			Name:  "launch",
			Usage: "launch a image as a new container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "runtime",
					Value: "",
				},
				cli.BoolFlag{
					Name: "nocache",
				},
				cli.BoolFlag{
					Name: "foreground, f",
				},
				cli.StringSliceFlag{
					Name:  "mount, m",
					Value: &cli.StringSlice{},
				},
				cli.BoolFlag{
					Name: "temporary, t",
				},
				cli.BoolFlag{
					Name: "nostart",
				},
			},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_launch)
			},
		},
		{
			Name:  "lxc",
			Usage: "lxc related commands",
			Subcommands: []cli.Command{
				{
					Name:  "config",
					Usage: "show container config",
					Flags: []cli.Flag{
						cli.StringFlag{
							Name:  "runtime",
							Value: "",
						},
					},
					Action: func(c *cli.Context) {
						run_cli(c, conf, cli_lxc_config)
					},
				},
			},
		},
		{
			Name:      "list",
			ShortName: "ls",
			Usage:     "list containers",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "runtime",
					Value: "",
				},
				cli.BoolFlag{
					Name: "all, a",
				},
			},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_list)
			},
		},
		{
			Name:  "start",
			Usage: "start a container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "runtime, r",
					Value: "",
				},
			},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_start)
			},
		},
		{
			Name:  "stop",
			Usage: "stop a container",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_stop)
			},
		},
		{
			Name:  "reboot",
			Usage: "reboot a container",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_reboot)
			},
		},
		{
			Name:  "shutdown",
			Usage: "shutdown a container",
			Flags: []cli.Flag{
				cli.DurationFlag{
					Name:  "timeout, t",
					Value: time.Duration(20) * time.Second,
					Usage: "how long to wait for container to shutdown",
				},
			},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_shutdown)
			},
		},
		{
			Name:  "freeze",
			Usage: "freeze a container",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_freeze)
			},
		},
		{
			Name:  "unfreeze",
			Usage: "unfreeze a container",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_unfreeze)
			},
		},
		{
			Name:  "destroy",
			Usage: "delete a container",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_destroy)
			},
		},
		{
			Name:  "rename",
			Usage: "rename a container",
			Flags: []cli.Flag{
				cli.BoolFlag{
					Name: "force, f",
				},
			},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_rename)
			},
		},
		{
			Name:  "info",
			Usage: "show generic info",
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_info)
			},
		},
		{
			Name:  "attach",
			Usage: "atach to running container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "namespaces",
					Value: "",
				},
				cli.StringFlag{
					Name:  "cwd",
					Value: "",
				},
				cli.IntFlag{
					Name: "uid, u",
				},
				cli.IntFlag{
					Name: "gid, g",
				},
				cli.BoolFlag{
					Name: "clear-env",
				},
				cli.BoolFlag{
					Name: "keep-env",
				},
			},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_attach)
			},
		},
		{
			Name:  "exec",
			Usage: "execute command in running container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "namespaces",
					Value: "",
				},
				cli.StringFlag{
					Name:  "cwd",
					Value: "",
				},
				cli.IntFlag{
					Name: "uid",
				},
				cli.IntFlag{
					Name: "gid",
				},
				cli.BoolFlag{
					Name: "clear-env",
				},
				cli.BoolFlag{
					Name: "keep-env",
				},
			},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_exec)
			},
		},
		{
			Name:  "import",
			Usage: "import a image",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "bootstrap",
					Value: "",
				},
			},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_import)
			},
		},
		{
			Name:  "quickstart",
			Usage: "quickly setup a cask directory template",
			Flags: []cli.Flag{},
			Action: func(c *cli.Context) {
				run_cli(c, conf, cli_quickstart)
			},
		},
	}

	app.Run(os.Args)

}

type CliFunc func(c *cli.Context, conf *config.Config)

func run_cli(c *cli.Context, conf *config.Config, fn CliFunc) {
	if conf.Sudo {
		current_user, err := user.Current()
		if err == nil {
			if current_user.Uid != "0" {
				sudo := "sudo"
				log.Tracef("exec %v %v", sudo, os.Args)
				cmd := exec.Command(sudo, os.Args...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = os.Stdout
				cmd.Stderr = os.Stderr
				err = cmd.Run()
				rc := 0
				if err != nil {
					rc = 1
				}
				os.Exit(rc)

			}
		} else {
			fmt.Println("current user: error getting current user:", err.Error())
		}
	}

	fn(c, conf)
}
