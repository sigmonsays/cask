package cgroup

import (
	"fmt"
	"gopkg.in/lxc/go-lxc.v2"
)

type MemoryBuilder struct {
	c *lxc.Container
}

func NewMemoryBuilder(c *lxc.Container) *MemoryBuilder {
	return &MemoryBuilder{
		c: c,
	}
}

func (b *MemoryBuilder) SetConfigItem(key, value string) error {
	log.Debugf("SetConfig %s = %s", key, value)
	err := b.c.SetConfigItem(key, value)
	if err != nil {
		log.Warnf("SetConfig %s = %s: %s", key, value, err)
	}
	return err
}

func (b *MemoryBuilder) Limit(bytes int) *MemoryBuilder {
	b.SetConfigItem("lxc.cgroup.memory.limit_in_bytes", fmt.Sprintf("%d", bytes))
	b.SetConfigItem("lxc.cgroup.memory.memsw.limit_in_bytes", fmt.Sprintf("%d", bytes))
	return b
}
