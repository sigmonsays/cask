package main

import (
	"encoding/json"
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/container"
	"github.com/sigmonsays/cask/image"
	"github.com/sigmonsays/cask/metadata"
	. "github.com/sigmonsays/cask/util"
	"github.com/termie/go-shutil"
	"gopkg.in/lxc/go-lxc.v2"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
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

func cli_build(c *cli.Context, conf *config.Config) {
	cmd := monitor()
	build_image(c, conf)
	cmd.Process.Signal(os.Interrupt)
	cmd.Wait()
}

func build_image(ctx *cli.Context, conf *config.Config) {

	opts := &BuildOptions{
		CommonOptions:  GetCommonOptions(ctx),
		keep_container: ctx.Bool("keep"),
		runtime:        ctx.String("runtime"),
		caskpath:       ctx.String("caskpath"),
	}

	log.Tracef("wait %v", opts.waitMask)
	log.Info("lxcpath", conf.StoragePath)
	log.Info("cask build runtime", opts.runtime)
	log.Info("cask path", opts.caskpath)

	metadatapath := filepath.Join(opts.caskpath, "meta.json")

	meta := &metadata.Meta{}
	err := meta.ReadFile(metadatapath)
	if err != nil {
		log.Error("meta ReadFile:", err)
		return
	}

	if len(opts.runtime) == 0 && len(meta.Runtime) > 0 {
		opts.runtime = meta.Runtime
		log.Info("Using runtime from metadata", opts.runtime)
	} else if opts.runtime == "" {
		opts.runtime = GetDefaultRuntime()
		log.Info("Using detected runtime from metadata", opts.runtime)
	}

	runtime_containerpath := filepath.Join(conf.StoragePath, opts.runtime)
	log.Tracef("runtime container path %s", runtime_containerpath)
	runtime, err := container.NewContainer(runtime_containerpath)
	if err != nil {
		log.Error("creating runtime:", err)
		return
	}
	err = runtime.C.LoadConfigFile(runtime.Path("config"))
	if err != nil {
		log.Error("runtime load config:", err)
		return
	}

	container_name := fmt.Sprintf("build-%d", os.Getpid())
	log.Tracef("temporary container name %s", container_name)
	clone_options := lxc.CloneOptions{
		Backend:  lxc.Aufs,
		Snapshot: true,
		// ConfigPath: conf.StoragePath,
	}
	err = runtime.C.Clone(container_name, clone_options)
	if err != nil {
		log.Error("clone:", err)
		return
	}

	// get the clone
	clonepath := filepath.Join(conf.StoragePath, container_name)
	clone, err := container.NewContainer(clonepath)
	if err != nil {
		log.Error("NewContainer:", err)
		return
	}

	// configure the clone
	clone.Build.Common()

	// configure the clones rootfs
	runtime_rootfs := runtime.Path("rootfs")
	clone_rootfs := clone.Path("delta")
	rootfs := fmt.Sprintf("aufs:%s:%s", runtime_rootfs, clone_rootfs)
	clone.Build.SetConfigItem("lxc.rootfs", rootfs)

	veth := container.DefaultVethType()
	veth.Name = "eth0"
	veth.Link = conf.Network.Bridge
	clone.Build.Network.AddInterface(veth)

	log.Tracef("save clone config %s", clone.Path("config"))
	err = clone.C.SaveConfigFile(clone.Path("config"))
	if err != nil {
		log.Errorf("SaveConfigFile: %s: %s", clone.Path("config"), err)
		return
	}

	if opts.keep_container == false {
		defer func() {
			clone.C.Destroy()
		}()
	}

	// parse the rootfs
	/*
		rootfs_values := clone.C.ConfigItem("lxc.rootfs")
		log.Tracef("clone has rootfs values %s", rootfs_values)
		if len(rootfs_values) == 0 {
			log.Error("cloned container:", container_name, "has no rootfs")
			return
		}
		if rootfs_values[0] == "" {
			log.Error("cloned container:", container_name, "has no rootfs")
			return
		}
		rootfs_tmp := strings.Split(rootfs_values[0], ":")

		if len(rootfs_tmp) > 2 {
			deltapath = rootfs_tmp[2]
			caskpath = filepath.Join(deltapath, "cask")
		} else {
			deltapath = rootfs_values[0]
			caskrootfs = filepath.Join(opts.caskpath, "rootfs")
			caskpath = filepath.Join(deltapath, "cask")
		}
		log.Tracef("deltapath %s", deltapath)
		log.Tracef("caskrootfs %s", caskrootfs)
		log.Tracef("caskpath %s", caskpath)
	*/
	deltapath := clone.Path("delta")
	caskrootfs := filepath.Join(opts.caskpath, "rootfs")
	caskpath := filepath.Join(deltapath, "cask")

	containerpath := filepath.Join(conf.StoragePath, meta.Name)
	rootfs_dir := filepath.Join(containerpath, "rootfs")
	metadata_path := filepath.Join(containerpath, "meta.json")
	archive_path := containerpath + ".tar.gz"

	// destroy existing data
	os.RemoveAll(containerpath)

	log.Debug("cask path", caskpath)
	log.Debug("container path", containerpath)
	log.Debug("archive path", archive_path)

	// add our script to the rootfs (temporary, we'll delete later)
	err = MergeTree(opts.caskpath, caskpath, 0)
	if err != nil {
		log.Error("MergeTree", opts.caskpath, "to", caskpath, err)
		return
	}

	container_path := func(subpath string) string {
		return filepath.Join(deltapath, subpath[1:])
	}

	os.MkdirAll(container_path("/cask/bin"), 0544)
	os.MkdirAll(filepath.Join(containerpath, "cask"), 0755)

	// set the configuration in metadata from the runtime
	keys := runtime.C.ConfigKeys()
	for _, key := range keys {
		values := runtime.C.ConfigItem(key)

		if len(values) == 0 {
			continue
		}

		for _, value := range values {
			if value == "" {
				continue
			}
			meta.SetConfigItem(key, value)
		}
	}

	// extract any images from the build
	for _, img := range meta.Build.Images {
		log.Debugf("adding image %s to container", img)
		image_archive, err := image.LocateImage(conf.StoragePath, img)
		if err != nil {
			log.Errorf("Unable to locate image %s: %s", img, err)
			return
		}

		err = UntarImage(image_archive, deltapath, opts.verbose)
		if err != nil {
			log.Errorf("Unable to extract image %s: %s", image_archive, err)
			return
		}
	}

	// walk the rootfs dir and add all the files into the destination rootfs
	offset := len(caskrootfs)
	log.Debug("copy rootfs", caskrootfs)
	newpath := ""
	walkfn := func(path string, info os.FileInfo, err error) error {
		if err != nil {
			log.Error("walk", path, err)
			return err
		}
		newpath = filepath.Join(deltapath, path[offset:])

		if info.IsDir() {
			os.MkdirAll(newpath, info.Mode())
			return nil
		} else if info.Mode().IsRegular() {
			err = shutil.CopyFile(path, newpath, false)
			if err != nil {
				log.Error("copy file", newpath, err)
			}
		} else {
			log.Warn("skipping", path)
		}
		return nil
	}
	err = filepath.Walk(caskrootfs, walkfn)
	if err != nil {
		log.Error("walk:", err)
		return
	}

	// copy the cask binary in the container
	cask_bin, err := findSelf()
	if err != nil {
		log.Errorf("cask binary not found: %s", err)
		return
	}
	log.Tracef("cask binary at %s", cask_bin)
	os.MkdirAll(container_path("/sbin"), 0755)
	err = CopyFile(cask_bin, container_path("/sbin/cask-init"), 0755)
	if err != nil {
		log.Errorf("copy %s -> %s: %s", cask_bin, container_path("/sbin/cask-init"), err)
		return
	}
	// TODO: dont hard code
	lxc_init := "/usr/sbin/init.lxc.static"
	err = CopyFile(lxc_init, container_path("/sbin/lxc-init"), 0755)
	if err != nil {
		log.Error("copy:", err)
		return
	}

	// start the container
	err = clone.C.Start()
	if err != nil {
		log.Error("starting cloned container:", err)
		return
	}

	log.Tracef("Waiting for RUNNING state..")
	clone.C.Wait(lxc.RUNNING, opts.waitTimeout)

	if opts.waitMask >= container.WaitMaskNetwork {
		log.Infof("container started, waiting %s for network..", opts.waitNetworkTimeout)
		// wait for it to startup and get network
		iplist, err := clone.C.WaitIPAddresses(opts.waitNetworkTimeout)
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
	exit_code, err := clone.C.RunCommandStatus(cmd, attach_options)
	if err != nil {
		log.Error("RunCommand", cmd, err)
		return
	}

	if exit_code != 0 {
		log.Error("bad exit code:", exit_code)
		return
	}

	// remove rootfs/cask path from container
	os.RemoveAll(caskpath)

	err = clone.C.Stop()
	if err != nil {
		log.Error("stop", err)
	}
	log.Info("stopped container with runtime delta at", deltapath)

	// simple file convention for container file system
	// rename the delta into rootfs, ie LXPATH / NAME / rootfs /
	// image metadata is at LXPATH / NAME / cask / meta.json

	os.MkdirAll(containerpath, 0644)

	new_meta_blob, err := json.MarshalIndent(meta, "", "   ")
	if err != nil {
		log.Error("marshal", err)
		return
	}

	ioutil.WriteFile(metadata_path, new_meta_blob, 0422)

	log.Debug("rename", deltapath, "->", rootfs_dir)
	err = os.Rename(deltapath, rootfs_dir)
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
				log.Error("merge", include_file, err)
				return
			}
		}
	}

	os.Mkdir(deltapath, 0644)

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

	archive_info, err := TarImage(archive_path, containerpath, opts.verbose)
	if err != nil {
		log.Error("tar:", archive_path, err)
		return
	}

	fmt.Printf("created archive %s, %d bytes\n", archive_path, archive_info.Size())
}
