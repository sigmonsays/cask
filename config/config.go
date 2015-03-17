package config

import (
	"bytes"
	"fmt"
	"launchpad.net/goyaml"
	"os"
)

var DefaultPath = "/etc/cask.yaml"

func DefaultConfigPath() string {
	return DefaultPath
}

type Config struct {
	StoragePath string
}

func DefaultConfig() *Config {
	return &Config{
		StoragePath: "/var/lib/cask",
	}
}

func (cfg *Config) FromFile(path string) error {
	err := cfg.LoadYaml(path)
	return err
}

func (c *Config) LoadYaml(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}

	b := bytes.NewBuffer(nil)
	_, err = b.ReadFrom(f)
	if err != nil {
		return err
	}

	if err := c.LoadYamlBuffer(b.Bytes()); err != nil {
		return err
	}

	if err := c.FixupConfig(); err != nil {
		return err
	}

	return nil
}
func (c *Config) LoadYamlBuffer(buf []byte) error {
	err := goyaml.Unmarshal(buf, c)
	if err != nil {
		return err
	}
	return nil
}

func (conf *Config) PrintConfig() {
	d, err := goyaml.Marshal(conf)
	if err != nil {
		fmt.Println("Marshal error", err)
		return
	}
	fmt.Println(string(d))
}

func (c *Config) FixupConfig() error {
	return nil
}
