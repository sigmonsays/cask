package util

import (
	"os"
	"os/exec"
)

func UntarImage(archive, containerpath string, verbose bool) error {
	log.Infof("extracting", archive, "in", containerpath)
	os.MkdirAll(containerpath, 0755)
	tar_flag := "-vzxf"
	if verbose == false {
		tar_flag = "-zxf"
	}
	cmdline := []string{"tar", "--strip-components=1", tar_flag, archive}
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = containerpath
	err := cmd.Run()
	if err != nil {
		log.Errorf("untar in %s) Command %s: %s", cmd.Dir, cmdline, err)
		return nil
	}
	return err
}
