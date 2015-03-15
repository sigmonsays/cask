package sup

import (
	"fmt"
	"os"
	"os/signal"
)

func Main() {
	fmt.Printf("cask init system started\n")
	signals := make(chan os.Signal, 1)
	signal.Notify(signals, os.Interrupt, os.Kill)
Dance:
	for {
		select {
		case sig := <-signals:
			log.Infof("Received %s", sig)
			break Dance
		}
	}
}
