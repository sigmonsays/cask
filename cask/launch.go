package main
import (
   "os"
   "os/exec"
   "fmt"
   "flag"
   "path/filepath"
   "io/ioutil"
   "encoding/json"
   "gopkg.in/lxc/go-lxc.v2"
)

type LaunchOptions struct {

   // be more verbose in some cases
   verbose bool

   // runtime name to build image in, ie "ubuntu12"
   runtime string

   // lxcpath where lxc config is stored, ie /var/lib/lxc
   lxcpath string

   // name of the container
   name string

}

func launch() {

   opts := &LaunchOptions{
   }

   f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
   f.BoolVar(&opts.verbose, "verbose", false, "be verbose")
   f.StringVar(&opts.runtime, "runtime", "", "specify runtime to use")
   f.StringVar(&opts.lxcpath, "lxcpath", lxc.DefaultConfigPath(), "Use specified container path")
   f.StringVar(&opts.name, "name", "", "specify container name")
   f.Parse(os.Args[2:])

   if opts.name == "" {
      opts.name = fmt.Sprintf("container-%d", os.Getpid())
   }
   if len(os.Args) < 3 {
      fmt.Println("ERROR: Need archive path")
      return
   }

   archivepath := f.Args()[0]
   fmt.Println("launch", opts.name, "using", archivepath)

   containerpath := filepath.Join(opts.lxcpath, opts.name)
   configpath := filepath.Join(containerpath, "config")
   metadatapath := filepath.Join(containerpath, "meta.json")
   rootfspath := filepath.Join(containerpath, "rootfs")
   mountpath := filepath.Join(containerpath, "fstab")

   fmt.Println("containerpath", containerpath)
   fmt.Println("metadata path", metadatapath)
   fmt.Println("rootfs path", rootfspath)

   os.MkdirAll(containerpath, 0755)
   cmdline := []string{ "tar", "--strip-components=1", "-vzxf", archivepath }
   cmd := exec.Command(cmdline[0], cmdline[1:]...)
   cmd.Dir = containerpath
   err := cmd.Run()
   if err != nil {
      fmt.Println("ERROR", err)
      return
   }

   meta := &Meta{}

   meta_blob, err := ioutil.ReadFile(metadatapath)
   if err != nil {
      fmt.Println("ERROR", err)
      return
   }

   err = json.Unmarshal(meta_blob, meta)
   if err != nil {
      fmt.Println("ERROR", err)
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

   container, err := lxc.NewContainer(opts.name, opts.lxcpath)
   if err != nil {
      fmt.Println("ERROR", err)
      return
   }

   if container.Defined() {
      container.Destroy()
   }

   container.ClearConfig()

   fstab, err := os.Create(mountpath)
   if err != nil {
      fmt.Println("ERROR", err)
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
         fmt.Println("ERROR", err)
         return
      }
   }
   
   fmt.Println("configured", opts.lxcpath, opts.name)
   

}
