package main

import (
	"github.com/codegangsta/cli"
	gologging "github.com/sigmonsays/go-logging"
	"gopkg.in/lxc/go-lxc.v2"
	"os"
	"time"
)

func WarnIf(err error) {
	log.Warnf("%s", err)
}

func main() {

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
			Name:  "list",
			Usage: "list containers",
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
	}
	app.Run(os.Args)
}
