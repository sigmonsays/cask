package main

import (
	"github.com/termie/go-shutil"
	"os"
	"os/exec"
)

func FileExists(path string) bool {
	if _, err := os.Stat(path); err != nil {
		return false
	}
	return true
}

func CopyFile(src, dst string, mode int) error {
	err := shutil.CopyFile(src, dst, false)
	if err != nil {
		return err
	}
	err = os.Chmod(dst, os.FileMode(mode))
	if err != nil {
		return err
	}
	return nil
}

func GetDefaultRuntime() string {
	if out, err := exec.Command("lsb_release", "-cs").Output(); err == nil {
		if len(out) > 0 {
			return string(out)
		}
	}
	return ""
}
