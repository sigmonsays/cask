package container

import (
	"github.com/sigmonsays/cask/container/cgroup"
	"gopkg.in/lxc/go-lxc.v2"
)

type ConfigBuilder struct {
	c       *lxc.Container
	Network *NetworkBuilder
	Mount   *MountBuilder
	FS      *FilesystemBuilder
	Cgroup  *cgroup.CgroupBuilder
}

func NewConfigBuilder(c *lxc.Container) *ConfigBuilder {
	return &ConfigBuilder{
		c:       c,
		Network: NewNetworkBuilder(c),
		Mount:   NewMountBuilder(c),
		FS:      NewFilesystemBuilder(c),
		Cgroup:  cgroup.NewCgroupBuilder(c),
	}
}

func (b *ConfigBuilder) SetConfigItem(key, value string) *ConfigBuilder {
	log.Tracef("%s = %s", key, value)
	err := b.c.SetConfigItem(key, value)
	if err != nil {
		log.Warnf("%s = %s: %s", key, value, err)
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
func (b *ConfigBuilder) Logging(logfile, loglevel string) *ConfigBuilder {
	b.SetConfigItem("lxc.loglevel", loglevel)
	b.SetConfigItem("lxc.logfile", logfile)
	return b
}

func (b *ConfigBuilder) RootFilesystem(runtimerootfs, rootfspath string) *ConfigBuilder {
	// TODO: select best implementation on the fly

	// prepare the root file system to use AUFS
	// fs := NewAufsFilesystem(runtimerootfs)
	// fs.AddLayer(rootfspath)

	// prepare the root file system to use overlayfs
	fs := NewOverlayFilesystem(runtimerootfs)
	fs.AddLayer(rootfspath)
	b.FS.SetRoot(fs)

	return b
}
