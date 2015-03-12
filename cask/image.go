package main

type Meta struct {
	Runtime string              `json:"runtime"`
	Name    string              `json:"name"`
	Config  map[string][]string `json:"config"`
}
