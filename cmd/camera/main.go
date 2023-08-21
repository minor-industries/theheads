package main

import (
	_ "embed"
	"github.com/minor-industries/theheads/camera"
	"github.com/minor-industries/theheads/camera/cfg"
	"github.com/minor-industries/theheads/common/dotenv"
	"github.com/vrischmann/envconfig"
)

//go:embed default.env
var defaultEnv string

func main() {
	err := dotenv.SetEnvFromContent(defaultEnv)
	if err != nil {
		panic(err)
	}

	env := &cfg.Cfg{}

	err = envconfig.Init(env)
	if err != nil {
		panic(err)
	}

	camera.Run(env)
}
