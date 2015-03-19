package container

import (
	"fmt"
	"gopkg.in/lxc/go-lxc.v2"
)

type MountBuilder struct {
	c *lxc.Container
}

func NewMountBuilder(c *lxc.Container) *MountBuilder {
	return &MountBuilder{
		c: c,
	}
}
func (b *MountBuilder) SetConfigItem(key, value string) error {
	log.Debugf("%s = %s", key, value)
	err := b.c.SetConfigItem(key, value)
	if err != nil {
		log.Warnf("%s = %s: %s", key, value, err)
	}
	return err
}

// bind mount a path in the container from the host
// src is a path on the host
// dst is a path _relative_ in the container under its lxc.rootfs, ie .Bind("/data", "data")
func (b *MountBuilder) Bind(src, dst string) *MountBuilder {
	m := defaultMount()
	m.Spec = src
	m.File = dst
	m.Type = "none"
	m.Options = "bind"
	b.SetConfigItem("lxc.mount.entry", m.String())
	return b
}

// holds a mount line as a structure
// format <spec> <file> <type> <options> <req> <pass>
//     ie /host /var/lib/lxc/container/path none bind 0 0
type mount struct {
	Spec    string `json:"spec"`
	File    string `json:"file"`
	Type    string `json:"type"`
	Options string `json:"options"`
	Req     int    `json:"req"`
	Pass    int    `json:"pass"`
}

func defaultMount() *mount {
	return &mount{
		Spec: "none",
		File: "none",
	}
}

func (m *mount) String() string {
	return fmt.Sprintf("%s %s %s %s %d %d", m.Spec, m.File, m.Type, m.Options, m.Req, m.Pass)
}
