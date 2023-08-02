package main

import (
	_ "embed"
	"github.com/cacktopus/theheads/boss"
	"github.com/cacktopus/theheads/boss/cfg"
	"github.com/cacktopus/theheads/common/discovery"
	"github.com/cacktopus/theheads/common/dotenv"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

//go:embed default.env
var defaultEnv string

func run() error {
	env := &cfg.Cfg{}

	err := dotenv.SetEnvFromContent(defaultEnv)
	if err != nil {
		panic(err)
	}

	err = envconfig.Init(env)
	if err != nil {
		return errors.Wrap(err, "init env")
	}

	boss.Run(env, discovery.NewSerf("127.0.0.1:7373"))
	return nil
}

func main() {
	err := run()
	if err != nil {
		panic(err)
	}
}
