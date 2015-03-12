package main

import (
	"fmt"
	"github.com/termie/go-shutil"
	"os"
	"os/exec"
	"path/filepath"
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

// copy a file tree from src to dst
func CopyTree(src, dst string) error {
	var err error
	offset := len(src)
	fmt.Println("copytree", src, "to", dst)
	var newpath string
	walkfn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("walk", path, err)
			return err
		}
		newpath = filepath.Join(src, path[offset:])

		if info.IsDir() {
			os.MkdirAll(newpath, info.Mode())
			return nil
		} else if info.Mode().IsRegular() {
			err = shutil.CopyFile(path, newpath, false)
			if err != nil {
				fmt.Println("ERROR copytree", newpath, err)
			}
		} else {
			fmt.Println("WARNING: skipping", path)
		}
		fmt.Println("copytree: copied", newpath)
		return nil
	}
	err = filepath.Walk(src, walkfn)
	if err != nil {
		fmt.Println("ERROR walk:", err)
		return err
	}
	return nil
}
