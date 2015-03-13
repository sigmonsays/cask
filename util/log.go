package util

import (
	gologging "github.com/sigmonsays/go-logging"
)

var log gologging.Logger

func init() {
	log = gologging.Register("util", func(newlog gologging.Logger) { log = newlog })
}
