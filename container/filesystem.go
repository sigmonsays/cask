package container

import (
	"gopkg.in/lxc/go-lxc.v2"
	"strings"
)

type FilesystemSetter interface {
	String() string
}

type FilesystemBuilder struct {
	c *lxc.Container
}

func NewFilesystemBuilder(c *lxc.Container) *FilesystemBuilder {
	return &FilesystemBuilder{
		c: c,
	}
}
func (b *FilesystemBuilder) SetConfigItem(key, value string) error {
	log.Debugf("%s = %s", key, value)
	err := b.c.SetConfigItem(key, value)
	if err != nil {
		log.Warnf("%s = %s: %s", key, value, err)
	}
	return err
}

func (b *FilesystemBuilder) SetRoot(fs FilesystemSetter) *FilesystemBuilder {
	b.c.ClearConfigItem("lxc.rootfs")
	b.SetConfigItem("lxc.rootfs", fs.String())
	return b
}

// AUFS file system
type AufsFilesystem struct {
	layers []string
}

func NewAufsFilesystem(path string) *AufsFilesystem {
	fs := &AufsFilesystem{
		layers: make([]string, 0),
	}
	fs.AddLayer(path)
	return fs
}

func (fs *AufsFilesystem) AddLayer(path string) {
	fs.layers = append(fs.layers, path)
}
func (fs *AufsFilesystem) String() string {
	return "aufs:" + strings.Join(fs.layers, ":")
}

// Overlay file system
type OverlayFilesystem struct {
	layers []string
}

func NewOverlayFilesystem(path string) *OverlayFilesystem {
	fs := &OverlayFilesystem{
		layers: make([]string, 0),
	}
	fs.AddLayer(path)
	return fs
}
func (fs *OverlayFilesystem) AddLayer(path string) {
	fs.layers = append(fs.layers, path)
}
func (fs *OverlayFilesystem) String() string {
	return "overlayfs:" + strings.Join(fs.layers, ":")
}
