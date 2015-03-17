package util

import (
	"os"
	"os/exec"
	"path/filepath"
)

func UntarImage(archive, containerpath string, verbose bool) error {
	log.Infof("extracting %s in %s", archive, containerpath)
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

func TarImage(archive, containerpath string, verbose bool) (os.FileInfo, error) {
	tar_flags := "zcf"
	if verbose {
		tar_flags = "vzcf"
	}

	dirname := filepath.Dir(containerpath)
	basename := filepath.Base(containerpath)

	cmdline := []string{"tar", tar_flags, archive, basename}

	log.Debug("tar command", cmdline)
	tar_cmd := exec.Command(cmdline[0], cmdline[1:]...)
	tar_cmd.Dir = dirname
	tar_cmd.Stdout = os.Stdout
	tar_cmd.Stderr = os.Stderr
	err := tar_cmd.Run()
	if err != nil {
		return nil, err
	}

	archive_info, err := os.Stat(archive)
	if err != nil {
		return nil, err
	}

	return archive_info, nil
}
