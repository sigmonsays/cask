package main

import (
	"fmt"
	"github.com/codegangsta/cli"
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

	CheckPrerequisites()

	app := cli.NewApp()
	app.Name = "cask"
	app.Usage = "manage container lifecycle"
	app.Version = "0.0.1"
	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:  "lxcpath",
			Value: lxc.DefaultConfigPath(),
		},
		cli.BoolFlag{
			Name: "verbose",
		},
		cli.StringFlag{
			Name:  "level, l",
			Value: "WARN",
		},
		cli.IntFlag{
			Name:  "wait",
			Usage: "wait max",
			Value: 2,
		},
		cli.DurationFlag{
			Name:  "wait-timeout, w",
			Value: time.Duration(3) * time.Second,
		},
		cli.DurationFlag{
			Name:  "net-timeout",
			Value: time.Duration(15) * time.Second,
		},
	}
	app.Before = func(c *cli.Context) error {
		gologging.SetLogLevel(c.String("level"))
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
				build(c)
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
				launch(c)
			},
		},
		{
			Name:  "config",
			Usage: "show container config",
			Flags: []cli.Flag{
				cli.StringFlag{
					Name:  "name, n",
					Value: "",
				},
				cli.StringFlag{
					Name:  "runtime",
					Value: "",
				},
			},
			Action: func(c *cli.Context) {
				config(c)
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
				list(c)
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
				start(c)
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
				stop(c)
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
				destroy(c)
			},
		},
		{
			Name:  "info",
			Usage: "show generic info",
			Action: func(c *cli.Context) {
				info(c)
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
			},
			Action: func(c *cli.Context) {
				attach(c)
			},
		},
	}
	app.Run(os.Args)
}
