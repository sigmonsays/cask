package builder

import (
	"gopkg.in/lxc/go-lxc.v2"
)

type ConfigBuilder struct {
	c       *lxc.Container
	Network *NetworkBuilder
}

func NewConfigBuilder(c *lxc.Container) *ConfigBuilder {
	return &ConfigBuilder{
		c:       c,
		Network: NewNetworkBuilder(c),
	}
}

func (b *ConfigBuilder) DropCap(cap string) *ConfigBuilder {
	b.c.SetConfigItem("lxc.cap.drop", cap)
	return b
}
func (b *ConfigBuilder) KeepCap(cap string) *ConfigBuilder {
	b.c.SetConfigItem("lxc.cap.keep", cap)
	return b
}
