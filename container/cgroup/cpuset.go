package cgroup

import (
	"gopkg.in/lxc/go-lxc.v2"
)

type CpuSetBuilder struct {
	c *lxc.Container
}

func NewCpuSetBuilder(c *lxc.Container) *CpuSetBuilder {
	return &CpuSetBuilder{
		c: c,
	}
}

func (b *CpuSetBuilder) SetConfigItem(key, value string) *CpuSetBuilder {
	log.Tracef("%s %s", key, value)
	err := b.c.SetConfigItem(key, value)
	if err != nil {
		log.Warnf("%s = %s: %s", key, value, err)
	}
	return b
}

func (b *CpuSetBuilder) CPUs(cpus string) *CpuSetBuilder {
	b.SetConfigItem("lxc.cgroup.cpuset.cpus", cpus)
	return b
}
