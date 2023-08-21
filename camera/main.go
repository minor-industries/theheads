package camera

import (
	"github.com/minor-industries/theheads/camera/cfg"
	"github.com/minor-industries/theheads/camera/floodlight"
	"github.com/minor-industries/theheads/common/broker"
	"github.com/minor-industries/theheads/common/util"
	"go.uber.org/zap"
)

func Run(env *cfg.Cfg) {
	logger, err := util.NewLogger(false)
	if err != nil {
		panic(err)
	}

	msgBroker := broker.NewBroker()
	go msgBroker.Start()

	errCh := make(chan error)
	go func() {
		panic(<-errCh)
	}()

	fl := floodlight.NewFloodlight(env.FloodlightPin)
	err = fl.Setup()
	if err != nil {
		logger.Error("error setting up floodlight", zap.Error(err))
	}

	c := NewCamera(
		logger,
		env,
		msgBroker,
		fl,
	)
	c.Setup()
	c.Run()
}
