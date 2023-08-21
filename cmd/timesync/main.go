package main

import (
	"github.com/minor-industries/theheads/common/discovery"
	"github.com/minor-industries/theheads/common/util"
	"github.com/minor-industries/theheads/timesync"
	"github.com/minor-industries/theheads/timesync/cfg"
	"github.com/vrischmann/envconfig"
)

func main() {
	logger, _ := util.NewLogger(true)

	env := cfg.Defaults
	err := envconfig.Init(&env)
	if err != nil {
		panic(err)
	}

	timesync.Run(logger, &env, discovery.NewSerf("127.0.0.1:7373"), nil)
}
