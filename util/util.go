package util

import (
	"fmt"
	"github.com/termie/go-shutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
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
// destination can exist and the files in src are merged with destination
func MergeTree(src, dst string, strip int) error {
	var err error
	offset := len(src)
	fmt.Println("MergeTree", src, "to", dst)
	var newpath string
	walkfn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println("walk", path, err)
			return err
		}

		tmp := strings.Split(path[offset:], "/")
		if strip < len(tmp) {
			newpath = filepath.Join(dst, strings.Join(tmp[strip:], "/"))
		} else {
			newpath = filepath.Join(dst, path[offset:])
		}

		os.MkdirAll(filepath.Dir(newpath), info.Mode())

		if info.IsDir() {
			os.MkdirAll(newpath, info.Mode())
			return nil
		} else if info.Mode().IsRegular() {
			err = CopyFile(path, newpath, int(info.Mode()))
			if err != nil {
				fmt.Println("ERROR copytree", newpath, err)
			}
		} else {
			fmt.Println("WARNING: skipping", path)
		}
		return nil
	}
	err = filepath.Walk(src, walkfn)
	if err != nil {
		fmt.Println("ERROR walk:", err)
		return err
	}
	return nil
}
