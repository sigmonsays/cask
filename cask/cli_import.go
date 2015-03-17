package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	. "github.com/sigmonsays/cask/util"
	"gopkg.in/lxc/go-lxc.v2"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type ImportOptions struct {
	*CommonOptions

	// name of the container
	name string

	// temporary container name used for import
	tmp_name string

	// bootstrap the image using a template
	bootstrap string
}

func cli_import(c *cli.Context, conf *config.Config) {

	opts := &ImportOptions{
		CommonOptions: GetCommonOptions(c),
		name:          c.String("name"),
		bootstrap:     c.String("bootstrap"),
	}

	if opts.name == "" {
		log.Error("container name required")
		return
	}
	opts.tmp_name = fmt.Sprintf("%s-%d", opts.name, os.Getpid())

	if len(opts.bootstrap) > 0 {

		bootstrap := strings.Split(opts.bootstrap, ".")
		if len(bootstrap) < 2 {
			log.Error("bootstrap must be in the format of DISTRIBUTION.RELEASE")
			return
		}

		distribution := bootstrap[0]
		release := bootstrap[1]

		cmdline := []string{
			"-t", "download",
			"-n", opts.name,
			"--",
			"-d", distribution,
			"-r", release,
		}
		cmd := exec.Command(cmdline[0], cmdline[1:]...)
		err := cmd.Run()
		if err != nil {
			log.Error("cmd:", cmdline, err)
			return
		}
	}

	container, err := lxc.NewContainer(opts.name, conf.StoragePath)
	if err != nil {
		log.Error("getting container", opts.name, err)
		return
	}

	// this is the new container
	container2, err := lxc.NewContainer(opts.tmp_name, conf.StoragePath)
	if err != nil {
		log.Error("getting container", opts.tmp_name, err)
		return
	}

	if container2.Running() {
		container2.Stop()
	}
	container2.Destroy()

	containerpath := filepath.Join(container.ConfigPath(), opts.name)
	rootfspath := filepath.Join(containerpath, "rootfs")

	log.Info(opts.name, "containerpath", containerpath)
	log.Info(opts.name, "rootfspath", rootfspath)

	// create the metadata for the image
	meta := &Meta{}
	meta.Name = opts.name
	keys := container.ConfigKeys()
	for _, key := range keys {
		values := container.ConfigItem(key)
		if len(values) == 0 {
			continue
		}

		for _, value := range values {
			if value == "" {
				continue
			}
			meta.SetConfigItem(key, value)
		}
	}
	log.Tracef("meta %+v", meta)

	containerpath2 := filepath.Join(container.ConfigPath(), opts.tmp_name)
	rootfspath2 := filepath.Join(containerpath2, "rootfs")
	metapath2 := filepath.Join(containerpath2, "meta.json")
	log.Tracef("building new rootfs at %s", rootfspath2)

	// copy the rootfs tree
	err = MergeTree(rootfspath, rootfspath2, 0)
	if err != nil {
		log.Error("merge", err)
		return
	}

	// save the metadata
	err = meta.WriteFile(metapath2)
	if err != nil {
		log.Error("meta save file", err)
		return
	}

	// create the archive
	archivepath := filepath.Join(container2.ConfigPath(), opts.name) + ".tar.gz"
	log.Debugf("Creating %s", archivepath)
	archive_info, err := TarImage(archivepath, containerpath2, opts.verbose)
	if err != nil {
		log.Error("tar:", archivepath, err)
		return
	}

	fmt.Printf("created archive %s, %d bytes\n", archivepath, archive_info.Size())

	// cleanup after ourselves
	// container2.Destroy()

}
