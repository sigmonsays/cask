package container

import (
	"github.com/sigmonsays/cask/config"
	"gopkg.in/lxc/go-lxc.v2"
	"strings"
)

type FilesystemSetter interface {
	String() string
	AddLayer(string)
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
	log.Debugf("SetConfigItem %s = %s", key, value)
	err := b.c.SetConfigItem(key, value)
	if err != nil {
		log.Warnf("SetConfigItem %s = %s: %s", key, value, err)
	}
	return err
}

func (b *FilesystemBuilder) SetRoot(fs FilesystemSetter) *FilesystemBuilder {
	b.c.ClearConfigItem("lxc.rootfs")
	b.SetConfigItem("lxc.rootfs", fs.String())
	log.Tracef("SetRoot %s", fs.String())
	return b
}

// creates a new file system depending on host configuration
func NewFileSystem(conf *config.Config, path string) FilesystemSetter {
	// TODO: pick backend based on whats available and configuration
	return NewOverlayFilesystem(path)
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
