package util

import (
	"os"
	"os/exec"
)

type TarOptions struct {
	Verbose bool

	// for UntarImage, specific path to extract from archive
	Path string
}

func UntarImage(archive, containerpath string, opts *TarOptions) error {
	log.Infof("extracting %s in %s", archive, containerpath)
	os.MkdirAll(containerpath, 0755)
	tar_flag := "-vzxf"
	if opts.Verbose == false {
		tar_flag = "-zxf"
	}
	cmdline := []string{"tar", tar_flag, archive}
	if opts.Path != "" {
		cmdline = append(cmdline, opts.Path)
	}
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

func TarImage(archive, containerpath string, opts *TarOptions) (os.FileInfo, error) {
	tar_flags := "zcf"
	if opts.Verbose {
		tar_flags = "vzcf"
	}

	cmdline := []string{"tar", tar_flags, archive, "cask", "rootfs"}

	log.Debug("tar command", cmdline)
	tar_cmd := exec.Command(cmdline[0], cmdline[1:]...)
	tar_cmd.Dir = containerpath
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
