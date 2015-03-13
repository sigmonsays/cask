package main

type Meta struct {
	Runtime string              `json:"runtime"`
	Name    string              `json:"name"`
	Config  map[string][]string `json:"config"`
	Data    map[string]string   `json:"data"`
	Build   BuildParams         `json:"build"`
}
type BuildParams struct {

	// glob patterns of files to exclude in the saved image
	Exclude []string
}
