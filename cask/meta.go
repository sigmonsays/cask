package main

// describe the metadata associated with an image and a container
// when its launched
type Meta struct {
	// name of the image
	Name string `json:"name"`

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

	Mount []MountConfig `json:"mount"`
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
	//TODO
}

type MountConfig struct {
	BindMount []string `json:"bind_mount"`
}
