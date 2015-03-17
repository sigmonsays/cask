package main

import (
	"os"
	"os/exec"
)

// find the absolute path of the executable
func findSelf() (string, error) {
	arg0 := os.Args[0]

	// always use absolute path if invoked as such..
	if len(arg0) > 0 && arg0[0] == '/' {
		return arg0, nil
	}

	// try cask from $PATH
	path, err := exec.LookPath("cask")
	if err == nil {
		return path, nil
	}

	return arg0, nil
}
