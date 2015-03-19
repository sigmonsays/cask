package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/container"
	"github.com/sigmonsays/cask/container/caps"
	"github.com/sigmonsays/cask/util"
	"github.com/termie/go-shutil"
	"gopkg.in/lxc/go-lxc.v2"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

type LaunchOptions struct {
	*CommonOptions

	// do not perform any caching (for downloading)
	nocache bool

	// name of the container
	name string

	// keep application in foreground
	foreground bool
}

type LaunchFunc func() error

type LaunchFunctions struct {
	list []LaunchFunc
}

func NewLaunchFunctions() *LaunchFunctions {
	return &LaunchFunctions{
		list: make([]LaunchFunc, 0),
	}
}
func (l *LaunchFunctions) Add(f LaunchFunc) {
	l.list = append(l.list, f)
}
func (l *LaunchFunctions) Execute() error {
	var err error
	for _, f := range l.list {
		err = f()
	}
	return err
}

func cli_launch(ctx *cli.Context, conf *config.Config) {

	opts := &LaunchOptions{
		CommonOptions: GetCommonOptions(ctx),
		name:          ctx.String("name"),
		nocache:       ctx.Bool("nocache"),
		foreground:    ctx.Bool("foreground"),
	}

	wait := GetWaitOptions(ctx)

	if opts.name == "" {
		opts.name = fmt.Sprintf("container-%d", os.Getpid())
	}
	if len(os.Args) < 3 {
		log.Error("Need archive path")
		return
	}

	// used to execute commands after the container has started
	post_launch := NewLaunchFunctions()

	// download the archive over HTTP if its a URL
	archive := ctx.Args().First()
	var archivepath string
	if strings.HasPrefix(archive, "http") {
		_, err := url.Parse(archive)
		if err != nil {
			log.Errorf("bad url: %s: %s", archive, err)
			return
		}
		suffix := ".tar.gz"
		archivepath = filepath.Join(conf.StoragePath, opts.name) + suffix

		if util.FileExists(archivepath) == false || opts.nocache == true {
			log.Info("downloading", archive, "to", archivepath)
			f, err := os.Create(archivepath)
			if err != nil {
				log.Errorf("bad url: %s: %s", archive, err)
				return
			}

			resp, err := http.Get(archive)
			if err != nil {
				log.Errorf("download: %s: %s", archive, err)
				return
			}
			io.Copy(f, resp.Body)
			resp.Body.Close()
			f.Close()
			log.Info("downloaded", archive)
		}
	} else {
		archivepath = archive
	}

	log.Info("launch", opts.name, "using", archivepath)

	containerpath := filepath.Join(conf.StoragePath, opts.name)
	logfile := filepath.Join(conf.StoragePath, opts.name) + ".log"
	caskpath := filepath.Join(containerpath, "cask")
	configpath := filepath.Join(containerpath, "config")
	metadatapath := filepath.Join(caskpath, "meta.json")
	rootfspath := filepath.Join(containerpath, "rootfs")
	hostnamepath := filepath.Join(rootfspath, "etc/hostname")
	mountpath := filepath.Join(containerpath, "fstab")

	container_path := func(subpath string) string {
		return filepath.Join(rootfspath, subpath[1:])
	}

	log.Debug("containerpath", containerpath)
	log.Debug("metadata path", metadatapath)
	log.Debug("rootfs path", rootfspath)

	if util.FileExists(archivepath) == false {
		log.Error("Archive not found:", archivepath)
		return
	}

	c, err := container.NewContainer(containerpath)
	if err != nil {
		log.Error("NewContainer", err)
		return
	}

	build := c.Build

	if c.C.Defined() {
		log.Info("destroying existing container", opts.name)

		if c.C.Running() {
			err := c.C.Stop()
			if err != nil {
				log.Warn("Stop", opts.name, err)
			}
		}
		c.C.Destroy()
	}

	err = util.UntarImage(archivepath, containerpath, opts.verbose)
	if err != nil {
		log.Errorf("UntarImage (in %s): %s\n", containerpath, err)
		return
	}

	// load the meta
	err = c.LoadMetadata()
	if err != nil {
		log.Errorf("load container meta: %s: %s", opts.name, err)
		return
	}
	meta := c.Meta

	if meta.Runtime == "" {
		log.Errorf("runtime can not be empty: %s", opts.name)
		return
	}
	log.Debug("runtime", meta.Runtime)

	runtimepath := filepath.Join(conf.StoragePath, meta.Runtime)
	runtime, err := container.NewContainer(runtimepath)
	if err != nil {
		log.Errorf("getting runtime container: %s", err)
		return
	}
	log.Debugf("runtime at %s", runtime.Path("/"))

	c.C.ClearConfig()

	if opts.verbose {
		c.C.SetVerbosity(lxc.Verbose)
	}

	// begin container configuration
	build.Common()

	// os.MkdirAll(container_path("/proc"), 0755)
	os.MkdirAll(container_path("/sys/fs/cgroup"), 0755)
	os.MkdirAll(container_path("/dev/pts"), 0755)
	os.MkdirAll(container_path("/dev/shm"), 0755)

	build.SetConfigItem("lxc.loglevel", container.LogTrace)
	build.SetConfigItem("lxc.logfile", logfile)

	os.MkdirAll(filepath.Dir(mountpath), 0755)

	if util.FileExists(mountpath) == false {
		fstab, err := os.Create(mountpath)
		if err != nil {
			log.Error("Create", mountpath, err)
			return
		}
		fstab.Close()
	}

	// specific configuration for this container
	build.SetConfigItem("lxc.utsname", opts.name)

	rootfs := fmt.Sprintf("aufs:%s/rootfs:%s", runtimepath, rootfspath)
	build.SetConfigItem("lxc.rootfs", rootfs)
	build.SetConfigItem("lxc.mount", mountpath)

	post_launch.Add(func() error {
		attach_options := lxc.DefaultAttachOptions
		cmd := []string{"ifconfig"}
		c.C.RunCommand(cmd, attach_options)
		return nil
	})

	// provision CPU shares and CPU sets
	log.Debugf("configure cgroups..")
	if len(meta.Cgroup.Cpu.CPU) > 0 {
		build.Cgroup.CpuSet.CPUs(meta.Cgroup.Cpu.CPU)
	}
	if len(meta.Cgroup.Cpu.Shares) > 0 {
		build.Cgroup.Cpu.Shares(meta.Cgroup.Cpu.Shares)
	}

	// provision Memory limit
	if meta.Cgroup.Memory.Limit > 0 {
		build.Cgroup.Memory.Limit(meta.Cgroup.Memory.Limit)
	}

	log.Debug("config network")
	veth := container.DefaultVethType()
	veth.Link = conf.Network.Bridge
	build.Network.AddInterface(veth)
	/*
		// NOTE: for static network to be configure we need to ensure no DHCP is running in the container!!
		err = os.MkdirAll(container_path("/etc/network"), 0755)
		WarnIf(err)
		err = ioutil.WriteFile(container_path("/etc/network/interfaces"), []byte("auto lo\niface lo inet loopback"), 0744)
		WarnIf(err)
		TODO: static IP example..
		NetworkConfig := &cc.NetworkConfig{
			IPv4: cc.IPv4Config{
				IP:      "192.168.7.55/24",
				Gateway: "192.168.7.1",
			},
		}
		build.Network.AddInterface(veth).WithNetworkConfig(NetworkConfig)
	*/

	if c.C.Defined() == false {
		log.Debug("container", opts.name, "not defined, creating..")

		err := c.C.SaveConfigFile(configpath)
		if err != nil {
			log.Error("SaveConfig", err)
			return
		}
	}

	// set hostname in container
	os.MkdirAll(rootfspath, 0755)
	os.MkdirAll(filepath.Join(rootfspath, "/etc"), 0755)
	ioutil.WriteFile(hostnamepath, []byte(opts.name), 0444)

	// hack alert...
	// make sure /etc/mtab exists
	ioutil.WriteFile(filepath.Join(rootfspath, "/etc/mtab"), []byte{}, 0444)

	log.Info("configured", conf.StoragePath, opts.name)

	// add our script to the rootfs (temporary, we'll delete later)
	err = shutil.CopyTree(caskpath, filepath.Join(rootfspath, "cask"), nil)
	if err != nil {
		log.Error("CopyTree", err)
		return
	}

	// Process any options
	host_mount := "/host"
	if meta.Options.HostMount {
		log.Debug("adding host mount", host_mount)
		if util.FileExists(host_mount) == false {
			os.MkdirAll(host_mount, 0755)
		}
		path := container_path(host_mount)
		if util.FileExists(path) == false {
			os.MkdirAll(path, 0755)
		}
		build.Mount.Bind(host_mount, path)
	}
	for _, bind_mount := range meta.Mount.BindMount {
		log.Debug("adding bind mount", bind_mount)
		if util.FileExists(bind_mount) == false {
			os.MkdirAll(bind_mount, 0755)
		}
		path := container_path(bind_mount)
		if util.FileExists(path) == false {
			os.MkdirAll(path, 0755)
		}
		build.Mount.Bind(bind_mount, path)
	}

	// always drop these - "sys_module", "mac_admin", "mac_override", "sys_time",
	default_drop := []string{
		cap.CAP_SYS_MODULE,
		cap.CAP_MAC_ADMIN,
		cap.CAP_MAC_OVERRIDE,
		cap.CAP_SYS_TIME,
	}
	for _, d := range default_drop {
		build.SetConfigItem("lxc.cap.drop", d)
	}

	// set auto start configuration
	if meta.AutoStart.Enable {
		s := meta.AutoStart
		build.SetConfigItem("lxc.start.auto", "1")
		build.SetConfigItem("lxc.start.delay", fmt.Sprintf("%d", s.Delay))
		build.SetConfigItem("lxc.start.order", fmt.Sprintf("%d", s.Order))
		for _, g := range s.Groups {
			build.SetConfigItem("lxc.start.group", g)
		}
	}

	// add/drop capabilities
	for _, cap_add := range meta.CapAdd {
		build.SetConfigItem("lxc.cap.add", cap_add)
	}
	for _, cap_drop := range meta.CapDrop {
		build.SetConfigItem("lxc.cap.drop", cap_drop)
	}

	// save the configuration
	err = c.C.SaveConfigFile(configpath)
	if err != nil {
		log.Error("SaveConfigFile", configpath, err)
		return
	}

	if opts.foreground {
		// cmdline is what we execute
		var cmdline []string
		if len(ctx.Args()) > 1 {
			cmdline = ctx.Args()[1:]
			log.Tracef("using command from cli %s", cmdline)
		} else if len(meta.DefaultCmd) > 0 {
			cmdline = strings.Split(meta.DefaultCmd, " ")
			log.Tracef("using command from meta.default_cmd %s", meta.DefaultCmd)
		}
		if len(cmdline) == 0 {
			log.Errorf("cmdline must be given with foreground option")
			return
		}
		// TODO: Figure out how to do this without using lxc-execute ...
		args := []string{
			"--rcfile", configpath,
			"--name", c.C.Name(),
			"--lxcpath", c.C.ConfigPath(),
			"--logpriority", "DEBUG",
			"--logfile", logfile,
			"--",
		}
		args = append(args, cmdline...)
		log.Tracef("exec Command %s", args)
		cmd := exec.Command("lxc-execute", args...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			exit_code := 255
			if xerr, ok := err.(*exec.ExitError); ok {
				if status, ok := xerr.Sys().(syscall.WaitStatus); ok {
					exit_code = status.ExitStatus()
				}
			}
			os.Exit(exit_code)
			return
		}
		os.Exit(0)
	}

	err = c.Start()
	if err != nil {
		log.Errorf("container start %s: %s", opts.name, err)
		return
	}

	err = c.WaitStart(wait)
	if err != nil {
		log.Error("container wait start", opts.name, err)
		return
	}

	// execute launch script now
	if util.FileExists(filepath.Join(rootfspath, "/cask/launch")) {
		attach_options := lxc.DefaultAttachOptions
		attach_options.ClearEnv = false
		cmdline := []string{"sh", "-c", "/cask/launch"}
		exit_code, err := c.C.RunCommandStatus(cmdline, attach_options)
		if err != nil {
			log.Error("RunCommandStatus", cmdline, err)
			return
		}
		if exit_code != 0 {
			log.Error("bad exit code:", cmdline, exit_code)
			return
		}

	}

	// post launch scripts
	err = post_launch.Execute()
	if err != nil {
		log.Error("post launch error", err)
		return
	}

	// if we want to remove the /cask path from the container...
	// os.RemoveAll(filepath.Join(rootfspath, "cask"))

}
