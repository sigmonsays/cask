package main

import (
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/metadata"
	"github.com/sigmonsays/cask/util"
	"os"
)

type QuickstartOptions struct {
	*CommonOptions
	name string
}

func cli_quickstart(ctx *cli.Context, conf *config.Config) {
	opts := &QuickstartOptions{
		CommonOptions: GetCommonOptions(ctx),
		name:          ctx.String("name"),
	}
	if opts.name == "" {
		opts.name = "example"
	}

	os.MkdirAll("cask/rootfs", 0755)

	meta := &metadata.Meta{
		Name:    opts.name,
		Runtime: "CHANGEME",
		Version: "0.0.1",
	}
	if util.FileExists("cask/meta.json") == false {
		err := meta.WriteFile("cask/meta.json")
		if err != nil {
			log.Errorf("Unable to write cask/meta.json: %s", err)
			return
		}
	}

}
