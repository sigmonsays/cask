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
