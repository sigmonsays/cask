package main
import (
   "os"
   "fmt"
   "flag"
   "gopkg.in/lxc/go-lxc.v2"
)

type LaunchOptions struct {

   // be more verbose in some cases
   verbose bool

   // runtime name to build image in, ie "ubuntu12"
   runtime string

   // lxcpath where lxc config is stored, ie /var/lib/lxc
   lxcpath string

}

func launch() {

   opts := &LaunchOptions{
   }

   f := flag.NewFlagSet(os.Args[0], flag.ExitOnError)
   f.BoolVar(&opts.verbose, "verbose", false, "be verbose")
   f.StringVar(&opts.runtime, "runtime", "", "specify runtime to use")
   f.StringVar(&opts.lxcpath, "lxcpath", lxc.DefaultConfigPath(), "Use specified container path")
   f.Parse(os.Args[2:])

   fmt.Println("launch", os.Args)

}
