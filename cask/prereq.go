package main

import (
	"fmt"
	"gopkg.in/lxc/go-lxc.v2"
	"strings"
)

func CheckPrerequisites() error {
	version := lxc.Version()

	if strings.HasPrefix(version, "1.1.") == false {
		return fmt.Errorf("version requirement not met: need atleast 1.1.x, have %s", version)
	}
	return nil
}
