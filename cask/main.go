package main
import (
   "os"
   "fmt"
)
func main() {

   cmd := ""
   if len(os.Args) > 1 {
      cmd = os.Args[1]
   }

   if cmd == "build" {
      build()
   } else if cmd == "launch" {
      launch()
   } else {
      fmt.Println("ERROR: Invalid command. try build or launch")
      fmt.Println("Usage:", os.Args[0], "[command] options")
      os.Exit(1)
   }
}

