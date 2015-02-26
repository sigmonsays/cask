package main
import (
   "os"
   "fmt"
   "flag"
   "path/filepath"
   "gopkg.in/lxc/go-lxc.v2"
)

type ConfigOptions struct {

   // be more verbose in some cases
   verbose bool

   // runtime name to build image in, ie "ubuntu12"
   runtime string

   // lxcpath where lxc config is stored, ie /var/lib/lxc
   lxcpath string

   // name of the container
   name string

}

func config() {

   opts := &ConfigOptions{
   }

   f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
   f.BoolVar(&opts.verbose, "verbose", false, "be verbose")
   f.StringVar(&opts.runtime, "runtime", "", "specify runtime to use")
   f.StringVar(&opts.lxcpath, "lxcpath", lxc.DefaultConfigPath(), "Use specified container path")
   f.Parse(os.Args[2:])

   runtimepath := filepath.Join(opts.lxcpath, opts.runtime)

   fmt.Println("runtime", opts.runtime)
   fmt.Println("runtimepath", runtimepath)

   runtime, err := lxc.NewContainer(opts.runtime, opts.lxcpath)
   if err != nil {
      fmt.Println("ERROR getting runtime container", err)
      return
   }

   fmt.Println("-- runtime configuration --")

   network := runtime.ConfigItem("lxc.network")
   fmt.Println("network", network)

   keys := runtime.ConfigKeys()
   for _, key := range keys {
      values := runtime.ConfigItem(key)
      fmt.Printf("#%s %#v\n", key, values)

      for _, value := range values {
         if value == "" {
            continue
         }
         fmt.Printf("%s = %s\n", key, value)
      }
   }
   

}
