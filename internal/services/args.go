package services

import (
	"sync"

	"github.com/alexflint/go-arg"
)

var argsOnce sync.Once
var myArgs *Args

type Args struct {
	Config string `arg:"-c" default:"./pairdump.yaml" help:"config file (.yaml)"`
}

func GetArgs() *Args {
	argsOnce.Do(func() {
		myArgs = &Args{}
		arg.MustParse(myArgs)
	})
	return myArgs
}
