package main
import (
   "os"
   "fmt"
)

func main() {

   commands := make(map[string]func(), 0)
   commands["build"] = build
   commands["launch"] = launch
   commands["config"] = config

   cmd := ""
   if len(os.Args) > 1 {
      cmd = os.Args[1]
   }

   cmdfun, ok := commands[cmd]

   if ok == false {
      fmt.Println("ERROR: Invalid command. try build or launch")
      fmt.Println("Usage:", os.Args[0], "[command] options")
      os.Exit(1)
      return
   }

   cmdfun()
}

