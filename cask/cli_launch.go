package main

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/sigmonsays/cask/config"
	"github.com/sigmonsays/cask/container"
	"github.com/sigmonsays/cask/image"
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

	// mounts
	mounts []string

	// destroy container when it exits
	temporary bool
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
		mounts:        ctx.StringSlice("mount"),
		temporary:     ctx.Bool("temporary"),
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
	if archive == "" {
		log.Errorf("archive argument rquired")
		return
	}
	var archivepath string
	if archive[0] == '/' {
		// its an absolute path
		archivepath = archive
	} else if strings.HasPrefix(archive, "http") {
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
		// relative path given to StoragePath
		image_archive, err := image.LocateImage(conf.StoragePath, archive)
		if err != nil {
			log.Errorf("LocateImage: %s: %s", archive, err)
			return
		}
		archivepath = image_archive
		log.Tracef("found image %s at %s", archive, archivepath)

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
	runtimerootfs := filepath.Join(runtimepath, "rootfs")
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
	c.Build.Logging(logfile, container.LogTrace)

	os.MkdirAll(container_path("/proc"), 0755)
	os.MkdirAll(container_path("/sys/fs/cgroup"), 0755)
	os.MkdirAll(container_path("/dev/pts"), 0755)
	os.MkdirAll(container_path("/dev/shm"), 0755)

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

	// setup root file system
	build.RootFilesystem(runtimerootfs, rootfspath)

	// prepare mounts
	build.SetConfigItem("lxc.mount", mountpath)

	/*
		TODO: add a post_launch set of commands that will be invoked
			post_launch.Add(func() error {
				attach_options := lxc.DefaultAttachOptions
				cmd := []string{"ifconfig"}
				c.C.RunCommand(cmd, attach_options)
				return nil
			})
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

	// configure any bind mounts from the cli
	for _, mount := range opts.mounts {
		if len(mount) < 1 {
			log.Warnf("skipping invalid mount %s", mount)
		}
		build.Mount.Bind(mount, mount[1:])
		os.MkdirAll(container_path(mount), 0755)
	}

	err = c.Prepare(conf, meta)
	if err != nil {
		log.Errorf("Prepare: %s", err)
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
		exit_code := 0
		if err != nil {
			exit_code = 255
			if xerr, ok := err.(*exec.ExitError); ok {
				if status, ok := xerr.Sys().(syscall.WaitStatus); ok {
					exit_code = status.ExitStatus()
				}
			}
		}

		if opts.temporary {
			err = c.C.Destroy()
			if err != nil {
				log.Errorf("destroying temporary container %s: %s", opts.name, err)
				return
			}
		}
		os.Exit(exit_code)
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
