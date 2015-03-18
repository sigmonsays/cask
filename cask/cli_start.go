package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	cc "github.com/sigmonsays/cask/container"
	"gopkg.in/lxc/go-lxc.v2"
	"path/filepath"
)

type StartOptions struct {
	*CommonOptions

	// name of the container
	name string
}

func cli_start(c *cli.Context, conf *config.Config) {

	opts := &StartOptions{
		CommonOptions: GetCommonOptions(c),
		name:          c.String("name"),
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

	// load the meta
	metadatapath := filepath.Join(conf.StoragePath, opts.name, "cask/meta.json")
	meta := NewMeta(opts.name)
	err = meta.ReadFile(metadatapath)
	if err != nil {
		log.Error("load container meta", opts.name, err)
		return
	}

	logfile := filepath.Join(conf.StoragePath, opts.name) + ".log"

	// update container paths with current storage path
	// reset specific config params
	reset_config := []string{
		"lxc.loglevel",
		"lxc.logfile",
		"lxc.rootfs",
		"lxc.mount",
	}

	for _, c := range reset_config {
		err = container.ClearConfigItem(c)
		if err != nil {
			log.Warnf("reset config %s: %s", c, err)
		}
	}

	// begin container configuration
	build := cc.NewConfigBuilder(container)
	build.SetConfigItem("lxc.loglevel", cc.LogTrace)
	build.SetConfigItem("lxc.logfile", logfile)

	runtimepath := filepath.Join(conf.StoragePath, meta.Runtime)
	rootfspath := filepath.Join(conf.StoragePath, opts.name, "rootfs")
	rootfs := fmt.Sprintf("aufs:%s:%s", runtimepath, rootfspath)
	build.SetConfigItem("lxc.rootfs", rootfs)
	build.SetConfigItem("lxc.mount", filepath.Join(conf.StoragePath, opts.name, "fstab"))

	configpath := filepath.Join(conf.StoragePath, opts.name, "config")

	log.Tracef("save config %s", configpath)
	err = container.SaveConfigFile(configpath)
	if err != nil {
		log.Error("SaveConfig", err)
		return
	}

	log.Info(opts.name, "is", container.State())

	err = container.Start()
	if err != nil {
		log.Error("starting container", opts.name, err)
		return
	}
}
