package main

import (
	_ "embed"
	"github.com/minor-industries/platform/common/discovery"
	"github.com/minor-industries/platform/common/dotenv"
	"github.com/minor-industries/theheads/boss"
	"github.com/minor-industries/theheads/boss/cfg"
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
