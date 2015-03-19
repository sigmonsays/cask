package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/sup"
	gologging "github.com/sigmonsays/go-logging"
	"gopkg.in/lxc/go-lxc.v2"
	"os"
	"path/filepath"
	"time"
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
				cli_build(c, conf)
			},
		},
		{
			Name:  "launch",
			Usage: "launch a image as a new container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "",
				},
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
			},
			Action: func(c *cli.Context) {
				cli_launch(c, conf)
			},
		},
		{
			Name:  "config",
			Usage: "show container config",
			Flags: []cli.Flag{
				cli.StringSliceFlag{
					Name:  "name, n",
					Value: &cli.StringSlice{},
				},
				cli.StringFlag{
					Name:  "runtime",
					Value: "",
				},
			},
			Action: func(c *cli.Context) {
				cli_config(c, conf)
			},
		},
		{
			Name:      "list",
			ShortName: "ls",
			Usage:     "list containers",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "",
				},
				cli.StringFlag{
					Name:  "runtime",
					Value: "",
				},
				cli.BoolFlag{
					Name: "all, a",
				},
			},
			Action: func(c *cli.Context) {
				cli_list(c, conf)
			},
		},
		{
			Name:  "start",
			Usage: "start a container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "",
				},
				cli.StringFlag{
					Name:  "runtime, r",
					Value: "",
				},
			},
			Action: func(c *cli.Context) {
				cli_start(c, conf)
			},
		},
		{
			Name:  "stop",
			Usage: "stop a container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "",
				},
			},
			Action: func(c *cli.Context) {
				cli_stop(c, conf)
			},
		},
		{
			Name:  "destroy",
			Usage: "delete a container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "",
				},
			},
			Action: func(c *cli.Context) {
				cli_destroy(c, conf)
			},
		},
		{
			Name:  "info",
			Usage: "show generic info",
			Action: func(c *cli.Context) {
				cli_info(c, conf)
			},
		},
		{
			Name:  "attach",
			Usage: "atach to running container",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "",
				},
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
				cli_attach(c, conf)
			},
		},
		{
			Name:  "import",
			Usage: "import a image",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "",
				},
				cli.StringFlag{
					Name:  "bootstrap",
					Value: "",
				},
			},
			Action: func(c *cli.Context) {
				cli_import(c, conf)
			},
		},
		{
			Name:  "quickstart",
			Usage: "quickly setup a cask directory template",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "",
				},
			},
			Action: func(c *cli.Context) {
				cli_quickstart(c, conf)
			},
		},
	}
	app.Run(os.Args)
}
