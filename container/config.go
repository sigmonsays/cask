package container

import (
	"github.com/sigmonsays/cask/container/cgroup"
	"gopkg.in/lxc/go-lxc.v2"
)

type ConfigBuilder struct {
	c       *lxc.Container
	Network *NetworkBuilder
	Memory  *MemoryBuilder
	Mount   *MountBuilder
	Cgroup  *cgroup.CgroupBuilder
}

func NewConfigBuilder(c *lxc.Container) *ConfigBuilder {
	return &ConfigBuilder{
		c:       c,
		Network: NewNetworkBuilder(c),
		Mount:   NewMountBuilder(c),
		Cgroup:  cgroup.NewCgroupBuilder(c),
	}
}

func (b *ConfigBuilder) SetConfigItem(key, value string) *ConfigBuilder {
	log.Tracef("SetConfigItem %s %s", key, value)
	err := b.c.SetConfigItem(key, value)
	if err != nil {
		log.Warnf("SetConfigItem %s = %s: %s", key, value, err)
	}
	return b
}

func (b *ConfigBuilder) DropCap(cap string) *ConfigBuilder {
	b.SetConfigItem("lxc.cap.drop", cap)
	return b
}
func (b *ConfigBuilder) KeepCap(cap string) *ConfigBuilder {
	b.SetConfigItem("lxc.cap.keep", cap)
	return b
}

// setup common configuration
func (b *ConfigBuilder) Common() *ConfigBuilder {

	params := map[string]string{
		"lxc.devttydir":   "lxc",
		"lxc.pts":         "1024",
		"lxc.tty":         "4",
		"lxc.pivotdir":    "lxc_putold",
		"lxc.mount.auto":  "cgroup:mixed proc:mixed sys:mixed",
		"lxc.mount.entry": "/sys/fs/fuse/connections sys/fs/fuse/connections none bind,optional 0 0",
		"lxc.seccomp":     "/usr/share/lxc/config/common.seccomp",
		"lxc.include":     "/usr/share/lxc/config/common.conf.d/",
	}
	for k, v := range params {
		b.SetConfigItem(k, v)
	}
	return b
}
