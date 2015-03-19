package container

import (
	"fmt"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/container/caps"
	"github.com/sigmonsays/cask/metadata"
	"github.com/sigmonsays/cask/util"
	"gopkg.in/lxc/go-lxc.v2"
	"os"
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
func (c *Container) Path(path string) string {
	return c.path(path)
}

func (c *Container) LoadMetadata() error {
	m := &metadata.Meta{}
	path := c.Path("/cask/meta.json")
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

	/*
			runtimepath := filepath.Join(filepath.Dir(c.path("/")), c.Meta.Runtime, "rootfs")
			rootfspath := c.path("/rootfs")
			rootfs := fmt.Sprintf("aufs:%s:%s", runtimepath, rootfspath)
		   c.Build.SetConfigItem("lxc.rootfs", rootfs)
		   c.Build.SetConfigItem("lxc.mount", c.path("/fstab"))
	*/

	// begin container configuration

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

// prepare the container according to its metadata
func (c *Container) Prepare(conf *config.Config, meta *metadata.Meta) error {

	container_path := func(s string) string {
		return filepath.Join(c.Path("/rootfs"), s)
	}
	build := c.Build

	// provision CPU shares and CPU sets
	log.Debugf("configure cgroups..")
	if len(meta.Cgroup.Cpu.CPU) > 0 {
		build.Cgroup.CpuSet.CPUs(meta.Cgroup.Cpu.CPU)
	}
	if len(meta.Cgroup.Cpu.Shares) > 0 {
		build.Cgroup.Cpu.Shares(meta.Cgroup.Cpu.Shares)
	}

	// provision Memory limit
	if meta.Cgroup.Memory.Limit > 0 {
		build.Cgroup.Memory.Limit(meta.Cgroup.Memory.Limit)
	}

	log.Debug("config network")
	veth := DefaultVethType()
	veth.Link = conf.Network.Bridge
	build.Network.AddInterface(veth)
	/*
		// NOTE: for static network to be configure we need to ensure no DHCP is running in the container!!
		err = os.MkdirAll(container_path("/etc/network"), 0755)
		WarnIf(err)
		err = ioutil.WriteFile(container_path("/etc/network/interfaces"), []byte("auto lo\niface lo inet loopback"), 0744)
		WarnIf(err)
		TODO: static IP example..
		NetworkConfig := &cc.NetworkConfig{
			IPv4: cc.IPv4Config{
				IP:      "192.168.7.55/24",
				Gateway: "192.168.7.1",
			},
		}
		build.Network.AddInterface(veth).WithNetworkConfig(NetworkConfig)
	*/

	// Process any options
	host_mount := "/host"
	if meta.Options.HostMount {
		log.Debug("adding host mount", host_mount)
		if util.FileExists(host_mount) == false {
			os.MkdirAll(host_mount, 0755)
		}
		path := container_path(host_mount)
		if util.FileExists(path) == false {
			os.MkdirAll(path, 0755)
		}
		build.Mount.Bind(host_mount, host_mount[1:])
	}
	for _, bind_mount := range meta.Mount.BindMount {
		log.Debug("adding bind mount", bind_mount)
		if util.FileExists(bind_mount) == false {
			os.MkdirAll(bind_mount, 0755)
		}
		path := container_path(bind_mount)
		if util.FileExists(path) == false {
			os.MkdirAll(path, 0755)
		}
		build.Mount.Bind(bind_mount, bind_mount[1:])
	}

	// always drop these - "sys_module", "mac_admin", "mac_override", "sys_time",
	default_drop := []string{
		caps.CAP_SYS_MODULE,
		caps.CAP_MAC_ADMIN,
		caps.CAP_MAC_OVERRIDE,
		caps.CAP_SYS_TIME,
	}
	for _, d := range default_drop {
		build.SetConfigItem("lxc.cap.drop", d)
	}

	// pass in environment variables
	for k, v := range meta.Env {
		build.SetConfigItem("lxc.environment", fmt.Sprintf("%s=%s", k, v))
	}

	// set auto start configuration
	if meta.AutoStart.Enable {
		s := meta.AutoStart
		build.SetConfigItem("lxc.start.auto", "1")
		build.SetConfigItem("lxc.start.delay", fmt.Sprintf("%d", s.Delay))
		build.SetConfigItem("lxc.start.order", fmt.Sprintf("%d", s.Order))
		for _, g := range s.Groups {
			build.SetConfigItem("lxc.start.group", g)
		}
	}

	// add/drop capabilities
	for _, cap_add := range meta.CapAdd {
		build.SetConfigItem("lxc.cap.add", cap_add)
	}
	for _, cap_drop := range meta.CapDrop {
		build.SetConfigItem("lxc.cap.drop", cap_drop)
	}

	// save the configuration
	configpath := c.Path("/config")
	err := c.C.SaveConfigFile(configpath)
	if err != nil {
		return err
	}
	return nil
}
