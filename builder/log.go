package builder

import (
	gologging "github.com/sigmonsays/go-logging"
)

var log gologging.Logger

func init() {
	log = gologging.Register("builder", func(newlog gologging.Logger) { log = newlog })
}
