package container

import (
	gologging "github.com/sigmonsays/go-logging"
)

var log gologging.Logger

func init() {
	log = gologging.Register("container", func(newlog gologging.Logger) { log = newlog })
}
