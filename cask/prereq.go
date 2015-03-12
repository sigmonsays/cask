package main
import (
   "fmt"
   "strings"
	"gopkg.in/lxc/go-lxc.v2"
)
func CheckPrerequisites() error {
   version := lxc.Version()
   fmt.Println("lxc version", version)

   if strings.HasPrefix(version, "1.1.") == false {
      return fmt.Errorf("version requirement not met: need atleast 1.1.x, have %s", version)
   }
   return nil
}
