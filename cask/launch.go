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
)

type LaunchOptions struct {
	*CommonOptions

	// runtime name to build image in, ie "ubuntu12"
	runtime string

	// name of the container
	name string
}

func launch(c *cli.Context) {

	opts := &LaunchOptions{
		CommonOptions: GetCommonOptions(c),
		name:          c.String("name"),
		runtime:       c.String("runtime"),
	}

	if opts.name == "" {
		opts.name = fmt.Sprintf("container-%d", os.Getpid())
	}
	if len(os.Args) < 3 {
		fmt.Println("ERROR: Need archive path")
		return
	}

	archivepath := c.Args().First()
	fmt.Println("launch", opts.name, "using", archivepath)

	containerpath := filepath.Join(opts.lxcpath, opts.name)
	caskpath := filepath.Join(containerpath, "cask")
	configpath := filepath.Join(containerpath, "config")
	metadatapath := filepath.Join(containerpath, "meta.json")
	rootfspath := filepath.Join(containerpath, "rootfs")
	hostnamepath := filepath.Join(rootfspath, "etc/hostname")
	mountpath := filepath.Join(containerpath, "fstab")

	fmt.Println("containerpath", containerpath)
	fmt.Println("metadata path", metadatapath)
	fmt.Println("rootfs path", rootfspath)

	if FileExists(archivepath) == false {
		fmt.Println("ERROR: Archive not found:", archivepath)
		return
	}

	container, err := lxc.NewContainer(opts.name, opts.lxcpath)
	if err != nil {
		fmt.Println("ERROR NewContainer", err)
		return
	}

	if container.Defined() {
		fmt.Println("destroying existing container", opts.name)
		err := container.Stop()
		if err != nil {
			fmt.Println("WARN Stop", opts.name, err)
		}
		container.Destroy()
	}

	fmt.Println("extracting", archivepath, "in", containerpath)
	os.MkdirAll(containerpath, 0755)
	tar_flag := "-vzxf"
	if opts.verbose == false {
		tar_flag = "-zxf"
	}
	cmdline := []string{"tar", "--strip-components=1", tar_flag, archivepath}
	cmd := exec.Command(cmdline[0], cmdline[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = containerpath
	err = cmd.Run()
	if err != nil {
		fmt.Printf("ERROR (in %s) Command %s: %s\n", cmd.Dir, cmdline, err)
		return
	}

	meta := &Meta{}

	meta_blob, err := ioutil.ReadFile(metadatapath)
	if err != nil {
		fmt.Println("ERROR ReadFile", metadatapath, err)
		return
	}

	err = json.Unmarshal(meta_blob, meta)
	if err != nil {
		fmt.Println("ERROR Unmarshal", err)
		return
	}

	fmt.Println("runtime", meta.Runtime)

	lxcruntimepath := filepath.Join(opts.lxcpath, meta.Runtime)
	runtime, err := lxc.NewContainer(meta.Runtime, opts.lxcpath)
	if err != nil {
		fmt.Println("ERROR getting runtime container", err)
		return
	}

	runtime_rootfs_values := runtime.ConfigItem("lxc.rootfs")
	runtimepath := runtime_rootfs_values[0]

	fmt.Println("runtime lxc config path", lxcruntimepath)
	fmt.Println("runtime path", runtimepath)

	container.ClearConfig()

	os.MkdirAll(filepath.Dir(mountpath), 0755)

	fstab, err := os.Create(mountpath)
	if err != nil {
		fmt.Println("ERROR Create", mountpath, err)
		return
	}

	// TODO: Add any additional mounts here
	fstab.Close()

	// merge config in from meta
	fmt.Println("merge config")
	for key, values := range meta.Config {

		if key == "lxc.network" {
			// work around for network configuration not being set properly
			continue
		}

		for _, value := range values {
			container.SetConfigItem(key, value)
		}
	}

	// specific configuration for this container
	container.SetConfigItem("lxc.utsname", opts.name)

	rootfs := fmt.Sprintf("aufs:%s:%s", runtimepath, rootfspath)
	container.SetConfigItem("lxc.rootfs", rootfs)
	container.SetConfigItem("lxc.mount", mountpath)

	fmt.Println("config network")
	container.SetConfigItem("lxc.network.type", "veth")
	container.SetConfigItem("lxc.network.link", "lxcbr0")
	container.SetConfigItem("lxc.network.flags", "up")
	container.SetConfigItem("lxc.network.hwaddr", "00:16:3e:xx:xx:xx")

	if container.Defined() == false {
		fmt.Println("container", opts.name, "not defined, creating..")

		err := container.SaveConfigFile(configpath)
		if err != nil {
			fmt.Println("ERROR SaveConfig", err)
			return
		}
	}

	os.MkdirAll(rootfspath, 0755)
	os.MkdirAll(filepath.Join(rootfspath, "/etc"), 0755)
	ioutil.WriteFile(hostnamepath, []byte(opts.name), 0444)

	fmt.Println("configured", opts.lxcpath, opts.name)

	// add our script to the rootfs (temporary, we'll delete later)
	err = shutil.CopyTree(caskpath, filepath.Join(rootfspath, "cask"), nil)
	if err != nil {
		fmt.Println("ERROR", err)
		return
	}

	// start the container
	err = container.Start()
	if err != nil {
		fmt.Println("ERROR Start", opts.name, err)
		return
	}

	// wait for container to start up....
	if opts.waitMask >= WaitMaskStart {
		container.Wait(lxc.RUNNING, opts.waitTimeout)
	}

	if opts.waitMask >= WaitMaskNetwork {
		fmt.Println("container started, waiting for network..")
		// wait for it to startup and get network
		iplist, err := container.WaitIPAddresses(opts.waitNetworkTimeout)
		fmt.Println("iplist", iplist)
		if err != nil {
			fmt.Println("WARNING did not get ip address from container", err)
		}
	}

	// execute launch script now
	if FileExists(filepath.Join(rootfspath, "/cask/launch")) {
		attach_options := lxc.DefaultAttachOptions
		attach_options.ClearEnv = false
		cmdline := []string{"sh", "-c", "/cask/launch"}
		exit_code, err := container.RunCommandStatus(cmdline, attach_options)
		if err != nil {
			fmt.Println("ERROR", cmdline, err)
			return
		}
		if exit_code != 0 {
			fmt.Println("ERROR bad exit code:", cmdline, exit_code)
			return
		}

	}
	// if we want to remove the /cask path from the container...
	// os.RemoveAll(filepath.Join(rootfspath, "cask"))

	fmt.Println("container", opts.name, "is running")
}
