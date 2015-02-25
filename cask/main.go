package main
import (
   "os"
   "time"
   "os/exec"
   "fmt"
   "flag"
   "path/filepath"
   "strings"
   "io/ioutil"
   "encoding/json"
   "gopkg.in/lxc/go-lxc.v2"
   "github.com/termie/go-shutil"
)
func main() {

   cmd := ""
   if len(os.Args) > 1 {
      cmd = os.Args[1]
   }

   if cmd == "build" {
      build()
   }
   
}

// how long to wait for the container to start (return RUNNING)
const waitTimeout = time.Duration(3) * time.Second

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

type Meta struct {
   Runtime string
   Name string
}

type BuildOptions struct {

   // be more verbose in some cases
   verbose bool

   // runtime name to build image in, ie "ubuntu12"
   runtime string

   // lxcpath where lxc config is stored, ie /var/lib/lxc
   lxcpath string

   // cask path where our rootfs and bootstrap script is found
   caskpath string

   // if we want to keep the build context container around after exit 
   keep_container bool
   waitNetworkTimeout time.Duration

}

func GetDefaultRuntime() string {
   if out, err := exec.Command("lsb_release", "-cs").Output(); err == nil {
      if len(out) > 0 {
         return string(out)
      }
   }
   return ""
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

func build() {
   cmd := monitor()
   build_image()
   cmd.Process.Signal(os.Interrupt)
   cmd.Wait()
}

func build_image() {

   cask_binary := os.Args[0]

   opts := &BuildOptions{
      caskpath: "cask",
      waitNetworkTimeout: time.Duration(1) * time.Second,
      keep_container: true,
   }

   f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
   f.BoolVar(&opts.verbose, "verbose", false, "be verbose")
   f.StringVar(&opts.runtime, "runtime", "", "specify runtime to use")
   f.StringVar(&opts.lxcpath, "lxcpath", lxc.DefaultConfigPath(), "Use specified container path")
   f.StringVar(&opts.caskpath, "caskpath", opts.caskpath, "override cask path")
   f.Parse(os.Args[2:])

   fmt.Println("lxcpath", opts.lxcpath)
   fmt.Println("cask build runtime", opts.runtime)
   fmt.Println("cask path", opts.caskpath)

   meta_path := filepath.Join(opts.caskpath, "meta.json")
   meta_blob, err := ioutil.ReadFile(meta_path)
   if err != nil {
      fmt.Println("ERROR meta.json:", err)
      return
   }

   meta := &Meta{}
   err = json.Unmarshal(meta_blob, meta)
   if err != nil {
      fmt.Println("ERROR meta.json:", err)
      return
   }

   if len(opts.runtime) == 0 && len(meta.Runtime) > 0 {
      opts.runtime = meta.Runtime
      fmt.Println("Using runtime from metadata", opts.runtime)
   } else if opts.runtime == "" {
      opts.runtime = GetDefaultRuntime()
      fmt.Println("Using detected runtime from metadata", opts.runtime)
   }
   

   runtime, err := lxc.NewContainer(opts.runtime, opts.lxcpath)
   if err != nil {
      fmt.Println("ERROR creating runtime:", err)
      return
   }

   container := fmt.Sprintf("build-%d", os.Getpid())
   clone_options := lxc.CloneOptions{
      Backend: lxc.Aufs,
      Snapshot: true,
   }
   err = runtime.Clone(container, clone_options)
   if err != nil {
      fmt.Println("ERROR clone:", err)
      return
   }
   
   // get the clone
   clone, err := lxc.NewContainer(container, opts.lxcpath)
   if err != nil {
      fmt.Println("ERROR clone NewContainer:", err)
      return
   }

   if opts.keep_container == false {
      defer clone.Destroy()
   }

   rootfs_values := clone.ConfigItem("lxc.rootfs")
   if len(rootfs_values) == 0 {
      fmt.Println("ERROR cloned container:", container, "has no rootfs")
      return
   }
   rootfs_tmp := strings.Split(rootfs_values[0], ":")
   delta_path := rootfs_tmp[2]
   cask_rootfs := filepath.Join(opts.caskpath, "rootfs")
   cask_path := filepath.Join(delta_path, "cask")

   fmt.Println("cask path", cask_path)

   // add our script to the rootfs
   err = shutil.CopyTree(opts.caskpath, cask_path, nil)
   if err != nil {
      fmt.Println("ERROR", err)
      return
   }

   container_path := func(subpath string) string {
      return filepath.Join(delta_path, subpath[1:])
   }

   os.MkdirAll(container_path("/cask/bin"), 0544)

   // copy the cask executable itself into the container
   err = CopyFile(cask_binary, container_path("/cask/bin/cask"), 0544)
   if err != nil {
      fmt.Println("ERROR", err)
      return
   }

   // walk the rootfs dir and add all the files into the destination rootfs

   offset := len(cask_rootfs)
   fmt.Println("copy rootfs", cask_rootfs)
   newpath := ""
   walkfn := func(path string, info os.FileInfo, err error) error {

      if err != nil {
         fmt.Println("walk", path, err)
         return err
      }
      newpath = filepath.Join(delta_path, path[offset:])

      if info.IsDir() {
         os.MkdirAll(newpath, info.Mode())
         return nil
      } else if info.Mode().IsRegular() {
         err = shutil.CopyFile(path, newpath, false)
         if err != nil {
            fmt.Println("ERROR copy file", newpath, err)
         }
      } else {
         fmt.Println("WARNING: skipping", path)
      }
      return nil
   }
   err = filepath.Walk(cask_rootfs, walkfn)
   if err != nil {
      fmt.Println("ERROR walk:", err)
      return
   }

   // start the container
   err = clone.Start()
   if err != nil {
      fmt.Println("ERROR starting cloned container:", err)
      return
   }

   clone.Wait(lxc.RUNNING, waitTimeout)

   // wait for it to startup and get network
   iplist, err := clone.WaitIPAddresses(time.Duration(opts.waitNetworkTimeout) * time.Second)

   fmt.Println("iplist", iplist, err)
   
   // execute bootstrap script now
   attach_options := lxc.DefaultAttachOptions
   attach_options.ClearEnv = false
   cmd := []string{ "sh", "-c", "/cask/bootstrap" }
   exit_code, err := clone.RunCommandStatus( cmd, attach_options)
   if err != nil {
      fmt.Println("ERROR", err)
      return
   }
   
   if exit_code != 0 {
      fmt.Println("ERROR bad exit code:", exit_code)
      return
   }


   err = clone.Stop()
   if err != nil {
      fmt.Println("ERROR stop", err)
   }
   fmt.Println("stopped container with runtime delta at", delta_path)

   archive_dir := filepath.Join(opts.lxcpath, meta.Name)
   rootfs_dir := filepath.Join(archive_dir, "rootfs")
   metadata_path := filepath.Join(archive_dir, "meta.json")
   archive_path := archive_dir + ".tar.gz"
   fmt.Println("archive dir", archive_dir)
   fmt.Println("archive path", archive_path)

   os.RemoveAll(archive_dir)

   // simple file convention for container file system
   // rename the delta into rootfs, ie LXPATH / NAME / rootfs / 
   // image metadata is at LXPATH / NAME / meta.json

   os.MkdirAll(archive_dir, 0644)

   ioutil.WriteFile(metadata_path, meta_blob, 0422)

   err = os.Rename(delta_path, rootfs_dir)
   if err != nil {
      fmt.Println("ERROR:", err)
      return
   }

   os.Mkdir(delta_path, 0644)

   // build a tar archive of the bugger

   tar_flags := "zcf"
   if opts.verbose {
      tar_flags = "vzcf"
   }

   cmdline :=  []string{ "tar", tar_flags, archive_path, meta.Name }

   fmt.Println("tar command", cmdline)
   tar_cmd := exec.Command(cmdline[0], cmdline[1:]...)
   tar_cmd.Dir = opts.lxcpath
   tar_cmd.Stdout = os.Stdout
   tar_cmd.Stderr = os.Stderr
   err = tar_cmd.Run()
   if err != nil {
      fmt.Println("ERROR tar", err)
      return
   }

   
}
