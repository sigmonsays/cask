package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/container"
	"github.com/sigmonsays/cask/metadata"
	"github.com/sigmonsays/cask/util"
)

type ImportOptions struct {
	*CommonOptions

	// name of the container
	name string

	// architecture
	arch string

	// temporary container name used for import
	tmp_name string

	// bootstrap the image using a template
	bootstrap string
}

func cli_import(c *cli.Context, conf *config.Config) {

	opts := &ImportOptions{
		CommonOptions: GetCommonOptions(c),
		name:          c.Args().First(),
		bootstrap:     c.String("bootstrap"),
		arch:          runtime.GOARCH,
	}

	if opts.name == "" {
		log.Error("container name required")
		return
	}

	importpath := c.Args().Get(1)
	if importpath == "" {
		log.Errorf("import path is required")
		return
	}
	log.Infof("importing %s rootfs from %s", opts.name, importpath)

	opts.tmp_name = fmt.Sprintf("%s-%d", opts.name, os.Getpid())

	if len(opts.bootstrap) > 0 {

		lxc_containerpath := fmt.Sprintf("/var/lib/lxc/%s", opts.name)
		if container, err := container.NewContainer(lxc_containerpath); err == nil {
			container.C.Destroy()
		}

		bootstrap := strings.Split(opts.bootstrap, ".")
		if len(bootstrap) < 2 {
			log.Error("bootstrap must be in the format of DISTRIBUTION.RELEASE")
			return
		}

		distribution := bootstrap[0]
		release := bootstrap[1]

		cmdline := []string{
			"lxc-create",
			"-t", "download",
			"-n", opts.name,
			"--",
			"-d", distribution,
			"-r", release,
			"-a", opts.arch,
		}
		cmd := exec.Command(cmdline[0], cmdline[1:]...)
		if opts.verbose {
			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr
		}
		err := cmd.Run()
		if err != nil {
			log.Error("cmd:", cmdline, err)
			return
		}
	}

	containerpath := filepath.Join(conf.StoragePath, opts.name)
	container, err := container.NewContainer(containerpath)
	if err != nil {
		log.Error("getting container", opts.name, err)
		return
	}

	if container.C.Running() {
		container.C.Stop()
	}
	container.C.Destroy()

	os.RemoveAll(containerpath)

	rootfspath := filepath.Join(containerpath, "rootfs")

	log.Info(opts.name, "containerpath", containerpath)
	log.Info(opts.name, "rootfspath", rootfspath)

	// configure the basic configuration
	container.Build.Common()
	container.Build.SetConfigItem("lxc.rootfs", rootfspath)

	// create the metadata for the image
	meta := &metadata.Meta{}
	meta.Name = opts.name
	meta.Runtime = opts.name
	keys := container.C.ConfigKeys()
	for _, key := range keys {
		values := container.C.ConfigItem(key)
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

	// copy the rootfs tree
	err = util.MergeTree(importpath, rootfspath, 0)
	if err != nil {
		log.Error("merge", err)
		return
	}

	// create the basic structure
	os.MkdirAll(container.Path("cask"), 0755)

	err = container.C.SaveConfigFile(container.Path("config"))
	if err != nil {
		log.Error("save container config", err)
		return
	}

	// save the metadata
	metapath := container.Path("cask/meta.json")
	err = meta.WriteFile(metapath)
	if err != nil {
		log.Error("meta save file", err)
		return
	}

	// create the archive
	topts := &util.TarOptions{
		Verbose: opts.verbose,
	}
	archivepath := filepath.Join(container.C.ConfigPath(), opts.name) + ".tar.gz"
	log.Debugf("Creating %s", archivepath)
	archive_info, err := util.TarImage(archivepath, containerpath, topts)
	if err != nil {
		log.Error("tar:", archivepath, err)
		return
	}

	fmt.Printf("created archive %s %d bytes\n", archivepath, archive_info.Size())

}
