package sup

import (
	gologging "github.com/sigmonsays/go-logging"
)

var log gologging.Logger

func init() {
	log = gologging.Register("cask-init", func(newlog gologging.Logger) { log = newlog })
}
