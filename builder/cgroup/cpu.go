package cgroup

import (
	"gopkg.in/lxc/go-lxc.v2"
	"time"
)

type CpuBuilder struct {
	c *lxc.Container
}

func NewCpuBuilder(c *lxc.Container) *CpuBuilder {
	return &CpuBuilder{
		c: c,
	}
}

func (b *CpuBuilder) SetConfigItem(key, value string) *CpuBuilder {
	log.Tracef("SetConfigItem %s %s", key, value)
	err := b.c.SetConfigItem(key, value)
	if err != nil {
		log.Warnf("SetConfigItem %s = %s: %s", key, value, err)
	}
	return b
}

func (b *CpuBuilder) Shares(shares string) *CpuBuilder {
	b.SetConfigItem("lxc.cgroup.cpu.shares", shares)
	return b
}

type CpuThrottler interface {
	Configure(*CpuBuilder) error
}

type CfsCpuThrottle struct {
	// period of time in microseconds
	Period time.Duration `json:"period"`

	// amount of time in period process can use (microseconds)
	Quota time.Duration `json:"quota"`
}

func (b *CpuBuilder) Throttle(throttler CpuThrottler) *CpuBuilder {
	err := throttler.Configure(b)
	if err != nil {
		log.Warnf("throttle %s", err)
	}
	return b
}
