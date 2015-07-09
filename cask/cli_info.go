package main

import (
	"fmt"
	"os/user"

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
	fmt.Println("sudo", conf.Sudo)

	current_user, err := user.Current()
	if err == nil {
		fmt.Println("current user", current_user.Uid, current_user.Username)
	} else {
		fmt.Println("current user: error getting current user:", err.Error())
	}

}
