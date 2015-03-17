package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"gopkg.in/lxc/go-lxc.v2"
)

type InfoOptions struct {
	*CommonOptions
}

func cli_info(c *cli.Context, conf *config.Config) {

	// opts := &InfoOptions{
	// 	CommonOptions: GetCommonOptions(c),
	// }

	fmt.Println("lxc version", lxc.Version())
	fmt.Println("lxc default config path", lxc.DefaultConfigPath())
	fmt.Println("storage path", conf.StoragePath)
}
