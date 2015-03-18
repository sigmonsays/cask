package container

import (
	"github.com/sigmonsays/cask/container/cgroup"
	"gopkg.in/lxc/go-lxc.v2"
)

type ConfigBuilder struct {
	c       *lxc.Container
	Network *NetworkBuilder
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
