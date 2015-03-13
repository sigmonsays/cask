package main

type Meta struct {
	// name of the image
	Name string `json:"name"`

	// version of the image
	Version string `json:"version"`

	// runtime this image should run with
	Runtime string `json:"runtime"`

	// holds configuration parameters such as lxc.*
	Config map[string][]string `json:"config"`

	// arbtrary data to be stored with image
	Data map[string]string `json:"data"`

	// build paramters for building the image
	Build BuildParams `json:"build"`
}
type BuildParams struct {

	// other images to include in this build
	Images []string `json:"images"`

	// glob patterns of files to exclude in the saved image
	Exclude []string `json:"exclude"`
}
