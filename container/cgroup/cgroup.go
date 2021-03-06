package cgroup

import (
	"gopkg.in/lxc/go-lxc.v2"
)

type CgroupBuilder struct {
	c      *lxc.Container
	CpuSet *CpuSetBuilder
	Cpu    *CpuBuilder
	Memory *MemoryBuilder
}

func NewCgroupBuilder(c *lxc.Container) *CgroupBuilder {
	return &CgroupBuilder{
		c:      c,
		CpuSet: NewCpuSetBuilder(c),
		Cpu:    NewCpuBuilder(c),
	}
}

func (b *CgroupBuilder) SetConfigItem(key, value string) *CgroupBuilder {
	log.Debugf("%s = %s", key, value)
	err := b.c.SetConfigItem(key, value)
	if err != nil {
		log.Warnf("%s = %s: %s", key, value, err)
	}
	return b
}
