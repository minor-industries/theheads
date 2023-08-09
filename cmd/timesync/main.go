package main

import (
	"github.com/cacktopus/theheads/common/discovery"
	"github.com/cacktopus/theheads/common/util"
	"github.com/cacktopus/theheads/timesync"
	"github.com/cacktopus/theheads/timesync/cfg"
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
