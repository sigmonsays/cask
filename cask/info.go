package main

import (
   "fmt"
	"github.com/codegangsta/cli"
	"gopkg.in/lxc/go-lxc.v2"
)

type InfoOptions struct {
	*CommonOptions
}

func info(c *cli.Context) {

	// opts := &InfoOptions{
	// 	CommonOptions: GetCommonOptions(c),
	// }

   fmt.Println("lxc version", lxc.Version())
   fmt.Println("lxc default config path", lxc.DefaultConfigPath())
}
