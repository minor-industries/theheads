package main

import (
	"github.com/minor-industries/platform/common/dotenv"
	"github.com/minor-industries/platform/common/util"
	"github.com/minor-industries/theheads/leds"
	"go.uber.org/zap"
)

func main() {
	logger, _ := util.NewLogger(false)

	dotenv.EnvOverrideFromFile(logger, "/boot/leds.env")

	err := leds.Run(logger)
	if err != nil {
		logger.Fatal("error", zap.Error(err))
	}
}
