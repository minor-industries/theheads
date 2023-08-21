package main

import (
	"github.com/minor-industries/theheads/head"
	"github.com/minor-industries/theheads/head/cfg"
	"github.com/vrischmann/envconfig"
)

func main() {
	env := &cfg.Cfg{}

	err := envconfig.Init(env)
	if err != nil {
		panic(err)
	}

	head.Run(env)
}
