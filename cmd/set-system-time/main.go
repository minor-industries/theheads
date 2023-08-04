package main

import (
	util2 "github.com/cacktopus/theheads/common/util"
	"github.com/cacktopus/theheads/timesync/rtc"
	"github.com/cacktopus/theheads/timesync/util"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"time"
)

func run(logger *zap.Logger) error {
	rtClock, err := rtc.SetupI2C()
	if err != nil {
		return errors.Wrap(err, "setup i2c")
	}

	err = rtClock.SetTime(time.Now())
	if err != nil {
		return errors.Wrap(err, "set time")
	}

	t, err := rtClock.ReadTime()
	if err != nil {
		return errors.Wrap(err, "read time")
	}

	logger.Info("setting system time", zap.String("t", t.String()))
	err = util.SetTime(t)
	if err != nil {
		return errors.Wrap(err, "set system time")
	}

	return nil
}

func main() {
	logger, _ := util2.NewLogger(false)

	if err := run(logger); err != nil {
		logger.Fatal("fatal", zap.Error(err))
	}
}
