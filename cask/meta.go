package main

import (
	"encoding/json"
	"io/ioutil"
)

// describe the metadata associated with an image and a container
// when its launched
type Meta struct {
	// name of the image
	Name string `json:"name"`

	// Default command to execute (optional)
	DefaultCmd string `json:"default_cmd,omitempty"`

	// Auto start
	AutoStart AutoStartConfig `json:"auto_start"`

	// version of the image
	Version string `json:"version"`

	// runtime this image should run with
	Runtime string `json:"runtime"`

	// holds configuration parameters such as lxc.*
	Config map[string][]string `json:"config"`

	// misc options
	Options Options `json:"options"`

	// arbtrary data to be stored with image
	Data map[string]string `json:"data"`

	// build paramters for building the image
	Build BuildParams `json:"build"`

	// network configuration
	Network NetworkConfig `json:"network"`

	// set capabilities
	CapAdd  []string `json:"cap_add"`
	CapDrop []string `json:"cap_drop"`

	// specific mount configuration
	Mount MountConfig `json:"mount"`

	// cgroup configuration
	Cgroup CgroupConfig `json:"cgroup"`
}

func (m *Meta) SetConfigItem(key, value string) error {
	if m.Config == nil {
		m.Config = make(map[string][]string)
	}
	if _, ok := m.Config[key]; ok == false {
		m.Config[key] = make([]string, 0)
	}
	m.Config[key] = append(m.Config[key], value)
	return nil
}

func (m *Meta) WriteFile(path string) error {
	buf, err := json.Marshal(m)
	if err != nil {
		return err
	}

	err = ioutil.WriteFile(path, buf, 0644)
	if err != nil {
		return err
	}
	return nil
}

type BuildParams struct {

	// other images to include in this build
	Images []string `json:"images"`

	// glob patterns of files to exclude in the saved image
	Exclude []string `json:"exclude"`
}

type Options struct {
	// setup host mount at /host when container launched
	HostMount bool `json:"host_mount"`
}

type NetworkConfig struct {
}

type MountConfig struct {
	BindMount []string `json:"bind_mount"`
}

type CgroupConfig struct {
	Cpu CpuConfig `json:"cpu"`
}

type CpuConfig struct {
	// how many shares the CPU can use
	Shares string `json:"shares"`

	// tie a specific process to given CPUs
	CPU string `json:"cpu"`
}

type AutoStartConfig struct {
	// enable autostart
	Enable bool `json:"enable"`

	// time to wait (seconds) after the container is started before starting the next
	Delay int `json:"delay"`

	// integer sorted to determine startup of containers
	Order int `json:"order"`

	// groups container belongs to
	Groups []string `json:"groups"`
}
