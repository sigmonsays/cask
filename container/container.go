package container

import (
	"fmt"
	"github.com/sigmonsays/cask/metadata"
	"gopkg.in/lxc/go-lxc.v2"
	"path/filepath"
)

type Container struct {

	// name of the container
	Name string

	// top directory of container, ie /var/lib/lxc/container-1
	directory string

	// metadata for container
	Meta *metadata.Meta

	// container handle
	C *lxc.Container

	// configuration tool
	Build *ConfigBuilder
}

func NewContainer(directory string) (*Container, error) {
	log.Tracef("NewContainer %s", directory)

	l, err := lxc.NewContainer(filepath.Base(directory), filepath.Dir(directory))
	if err != nil {
		return nil, err
	}
	l.ClearConfig()

	c := &Container{
		Name:      filepath.Base(directory),
		directory: directory,
		C:         l,
		Build:     NewConfigBuilder(l),
	}

	c.LoadMetadata()

	return c, nil
}

func (c *Container) path(path string) string {
	return filepath.Join(c.directory, path)
}

func (c *Container) LoadMetadata() error {
	m := &metadata.Meta{}
	path := c.path("/meta.json")
	log.Tracef("read metadata %s", path)
	err := m.ReadFile(path)
	if err != nil {
		return err
	}

	log.Tracef("meta %+v", m)
	c.Meta = m
	return nil
}

func (c *Container) mergeConfig() error {
	if c.Meta == nil {
		return fmt.Errorf("Metadata is nil")
	}
	log.Debug("merge lxc config")
	// merge lxc config in from metadata
	for key, values := range c.Meta.Lxc {

		if key == "lxc.network" {
			// work around for network configuration not being set properly
			continue
		}
		if key == "lxc.cgroup" {
			// work around for cgroup configuration not being set properly
			continue
		}

		for _, value := range values {
			c.Build.SetConfigItem(key, value)
		}
	}
	return nil
}

func (c *Container) Start() error {
	err := c.mergeConfig()
	if err != nil {
		return err
	}

	logfile := c.path("") + ".log"

	// update container paths with current storage path
	// reset specific config params
	/*
		reset_config := []string{
			"lxc.loglevel",
			"lxc.logfile",
			"lxc.rootfs",
			"lxc.mount",
		}
		for _, name := range reset_config {
			err = container.ClearConfigItem(name)
			if err != nil {
				log.Warnf("reset config %s: %s", c, err)
			}
		}
	*/

	// we always need to set the paths so containers are "relocatable" -- that is they can be moved around
	// on disk and started

	runtimepath := filepath.Join(filepath.Dir(c.path("/")), c.Meta.Runtime, "rootfs")
	rootfspath := c.path("/rootfs")
	rootfs := fmt.Sprintf("aufs:%s:%s", runtimepath, rootfspath)

	// begin container configuration
	c.Build.SetConfigItem("lxc.loglevel", LogTrace)
	c.Build.SetConfigItem("lxc.logfile", logfile)

	c.Build.SetConfigItem("lxc.rootfs", rootfs)
	c.Build.SetConfigItem("lxc.mount", c.path("/fstab"))

	configpath := c.path("config")
	log.Tracef("save config %s", configpath)
	err = c.C.SaveConfigFile(configpath)
	if err != nil {
		return err
	}

	log.Info(c.Name, "is", c.C.State())

	err = c.C.Start()
	if err != nil {
		return err
	}

	return nil
}

// wait for a container to start up properly
func (c *Container) WaitStart(opts *WaitOptions) error {

	container := c.C

	// wait for container to start up....
	if opts.WaitMask >= WaitMaskStart {
		log.Tracef("Waiting for container state RUNNING")
		container.Wait(lxc.RUNNING, opts.WaitTimeout)
	}

	var ip string
	var iplist []string
	if opts.WaitMask >= WaitMaskNetwork {
		log.Info("container started, waiting for network..")
		// wait for it to startup and get network
		iplist, err := container.WaitIPAddresses(opts.WaitNetworkTimeout)
		log.Debug("iplist", iplist)
		if err != nil {
			log.Warn("did not get ip address from container", err)
		}
	}
	if len(iplist) > 0 {
		ip = iplist[0]
	}
	if len(ip) > 0 {
		log.Info("container", c.Name, "is running with ip", ip)
	} else {
		log.Info("container", c.Name, "is running")
	}

	return nil
}
