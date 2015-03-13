package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/termie/go-shutil"
	"gopkg.in/lxc/go-lxc.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type BuildOptions struct {
	*CommonOptions

	// runtime name to build image in, ie "ubuntu12"
	runtime string

	// cask path where our rootfs and bootstrap script is found
	caskpath string

	// if we want to keep the build context container around after exit
	keep_container bool
}

func monitor() *exec.Cmd {
	cmd := exec.Command("lxc-monitor")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	err := cmd.Start()
	if err != nil {
		fmt.Println("ERROR: lxc-monitor", err)
		return nil
	}
	return cmd
}

func build(c *cli.Context) {
	cmd := monitor()
	build_image(c)
	cmd.Process.Signal(os.Interrupt)
	cmd.Wait()
}

func build_image(c *cli.Context) {

	opts := &BuildOptions{
		CommonOptions:  GetCommonOptions(c),
		keep_container: c.Bool("keep"),
		runtime:        c.String("runtime"),
		caskpath:       c.String("caskpath"),
	}

	log.Tracef("wait %v", opts.waitMask)
	log.Info("lxcpath", opts.lxcpath)
	log.Info("cask build runtime", opts.runtime)
	log.Info("cask path", opts.caskpath)

	meta_path := filepath.Join(opts.caskpath, "meta.json")
	meta_blob, err := ioutil.ReadFile(meta_path)
	if err != nil {
		log.Errorf("ERROR meta.json: %s", err)
		return
	}

	meta := &Meta{}
	err = json.Unmarshal(meta_blob, meta)
	if err != nil {
		log.Error("ERROR meta.json:", err)
		return
	}
	log.Tracef("meta %+v", meta)

	if meta.Config == nil {
		meta.Config = make(map[string][]string, 0)
	}

	if len(opts.runtime) == 0 && len(meta.Runtime) > 0 {
		opts.runtime = meta.Runtime
		log.Error("Using runtime from metadata", opts.runtime)
	} else if opts.runtime == "" {
		opts.runtime = GetDefaultRuntime()
		log.Error("Using detected runtime from metadata", opts.runtime)
	}

	runtime, err := lxc.NewContainer(opts.runtime, opts.lxcpath)
	if err != nil {
		log.Error("ERROR creating runtime:", err)
		return
	}

	container := fmt.Sprintf("build-%d", os.Getpid())
	log.Tracef("temporary container name %s", container)
	clone_options := lxc.CloneOptions{
		Backend:  lxc.Aufs,
		Snapshot: true,
	}
	err = runtime.Clone(container, clone_options)
	if err != nil {
		log.Error("ERROR clone:", err)
		return
	}

	// get the clone
	clone, err := lxc.NewContainer(container, opts.lxcpath)
	if err != nil {
		log.Error("ERROR clone NewContainer:", err)
		return
	}

	if opts.keep_container == false {
		defer func() {
			clone.Destroy()
		}()
	}

	rootfs_values := clone.ConfigItem("lxc.rootfs")
	if len(rootfs_values) == 0 {
		log.Error("ERROR cloned container:", container, "has no rootfs")
		return
	}
	rootfs_tmp := strings.Split(rootfs_values[0], ":")
	log.Tracef("clone has rootfs %s", rootfs_tmp)

	var delta_path string
	var cask_rootfs string
	var cask_path string
	if len(rootfs_tmp) > 2 {
		delta_path = rootfs_tmp[2]
		cask_rootfs = filepath.Join(opts.caskpath, "rootfs")
		cask_path = filepath.Join(delta_path, "cask")
	} else {
		delta_path = rootfs_values[0]
		cask_rootfs = filepath.Join(opts.caskpath, "rootfs")
		cask_path = filepath.Join(delta_path, "cask")
	}

	containerpath := filepath.Join(opts.lxcpath, meta.Name)
	rootfs_dir := filepath.Join(containerpath, "rootfs")
	metadata_path := filepath.Join(containerpath, "meta.json")
	archive_path := containerpath + ".tar.gz"

	// destroy existing data
	os.RemoveAll(containerpath)

	log.Debug("cask path", cask_path)
	log.Debug("container path", containerpath)
	log.Debug("archive path", archive_path)

	// add our script to the rootfs (temporary, we'll delete later)
	err = shutil.CopyTree(opts.caskpath, cask_path, nil)
	if err != nil {
		log.Error("ERROR CopyTree", opts.caskpath, "to", cask_path, err)
		return
	}

	container_path := func(subpath string) string {
		return filepath.Join(delta_path, subpath[1:])
	}

	os.MkdirAll(container_path("/cask/bin"), 0544)

	// save a copy of the config in the container
	// copy the runtime configuration
	os.MkdirAll(filepath.Join(containerpath, "cask"), 0755)

	fh, err := os.Create(filepath.Join(containerpath, "cask", "container-config"))
	if err != nil {
		log.Error("ERROR Create", err)
		return
	}
	keys := runtime.ConfigKeys()
	for _, key := range keys {
		values := runtime.ConfigItem(key)

		if len(values) == 0 {
			continue
		}

		if _, ok := meta.Config[key]; ok == false {
			meta.Config[key] = make([]string, 0)
		}

		for _, value := range values {
			if value == "" {
				continue
			}
			fmt.Fprintf(fh, "%s = %s\n", key, value)
			meta.Config[key] = append(meta.Config[key], value)
		}
	}
	fh.Close()

	// walk the rootfs dir and add all the files into the destination rootfs
	offset := len(cask_rootfs)
	log.Debug("copy rootfs", cask_rootfs)
	newpath := ""
	walkfn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error("walk", path, err)
			return err
		}
		newpath = filepath.Join(delta_path, path[offset:])

		if info.IsDir() {
			os.MkdirAll(newpath, info.Mode())
			return nil
		} else if info.Mode().IsRegular() {
			err = shutil.CopyFile(path, newpath, false)
			if err != nil {
				log.Error("ERROR copy file", newpath, err)
			}
		} else {
			log.Error("WARNING: skipping", path)
		}
		return nil
	}
	err = filepath.Walk(cask_rootfs, walkfn)
	if err != nil {
		log.Error("ERROR walk:", err)
		return
	}

	// start the container
	err = clone.Start()
	if err != nil {
		log.Error("ERROR starting cloned container:", err)
		return
	}

	clone.Wait(lxc.RUNNING, opts.waitTimeout)

	if opts.waitMask >= WaitMaskNetwork {
		log.Infof("container started, waiting %s for network..", opts.waitNetworkTimeout)
		// wait for it to startup and get network
		iplist, err := clone.WaitIPAddresses(opts.waitNetworkTimeout)
		if err != nil {
			log.Infof("WARNING did not get ip address from container: %s", err)
		}
		for _, ip := range iplist {
			fmt.Println("ip", ip)
		}
	}

	// execute bootstrap script now
	attach_options := lxc.DefaultAttachOptions
	attach_options.ClearEnv = false
	cmd := []string{"sh", "-c", "/cask/bootstrap"}
	exit_code, err := clone.RunCommandStatus(cmd, attach_options)
	if err != nil {
		log.Error("ERROR RunCommand", cmd, err)
		return
	}

	if exit_code != 0 {
		log.Error("ERROR bad exit code:", exit_code)
		return
	}

	// remove rootfs/cask path from container
	os.RemoveAll(cask_path)

	err = clone.Stop()
	if err != nil {
		log.Error("ERROR stop", err)
	}
	log.Info("stopped container with runtime delta at", delta_path)

	// simple file convention for container file system
	// rename the delta into rootfs, ie LXPATH / NAME / rootfs /
	// image metadata is at LXPATH / NAME / meta.json

	os.MkdirAll(containerpath, 0644)

	new_meta_blob, err := json.MarshalIndent(meta, "", "   ")
	if err != nil {
		log.Error("ERROR marshal", err)
		return
	}

	ioutil.WriteFile(metadata_path, new_meta_blob, 0422)

	log.Debug("rename", delta_path, "->", rootfs_dir)
	err = os.Rename(delta_path, rootfs_dir)
	if err != nil {
		log.Error("Rename:", err)
		return
	}

	// copy our cask into the container path next to rootfs
	included_files := []string{
		"meta.json",
		"launch",
		"bootstrap",
	}
	for _, filename := range included_files {
		include_file := filepath.Join(opts.caskpath, filename)
		if FileExists(include_file) {
			err = MergeTree(include_file, filepath.Join(containerpath, "cask", filename), 0)
			if err != nil {
				log.Error("MergeTree", include_file, err)
				return
			}
		}
	}

	os.Mkdir(delta_path, 0644)

	// process image exclusions
	for _, exclude := range meta.Build.Exclude {
		log.Tracef("processing image exclusion %s for container path %s", exclude, containerpath)
		matches, err := filepath.Glob(containerpath + "/rootfs/" + exclude)
		if err != nil {
			log.Warnf("glob [%s] error %s", exclude, err)
			continue
		}
		for _, match := range matches {
			log.Tracef("delete %s", match)
			err := os.RemoveAll(match)
			if err != nil {
				log.Warnf("RemoveAll %s: %s", match, err)
			}
		}
	}

	// build a tar archive of the bugger

	tar_flags := "zcf"
	if opts.verbose {
		tar_flags = "vzcf"
	}

	cmdline := []string{"tar", tar_flags, archive_path, meta.Name}

	log.Debug("tar command", cmdline)
	tar_cmd := exec.Command(cmdline[0], cmdline[1:]...)
	tar_cmd.Dir = opts.lxcpath
	tar_cmd.Stdout = os.Stdout
	tar_cmd.Stderr = os.Stderr
	err = tar_cmd.Run()
	if err != nil {
		log.Error("ERROR tar", err)
		return
	}

	archive_info, err := os.Stat(archive_path)
	if err != nil {
		log.Error("failed to build archive:", archive_path, err)
		return
	}

	fmt.Printf("created archive %s, %d bytes\n", archive_path, archive_info.Size())
}
