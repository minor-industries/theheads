package main

import (
	"fmt"
	util2 "github.com/cacktopus/theheads/common/util"
	"github.com/cacktopus/theheads/timesync/rtc"
	"github.com/cacktopus/theheads/timesync/rtc/ds3231"
	"github.com/cacktopus/theheads/timesync/util"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
)

var opts = struct {
	Show struct {
	} `command:"show"`

	SetSystem struct {
	} `command:"set-system"`
}{}

func run(logger *zap.Logger) error {
	parser := flags.NewParser(&opts, flags.Default)
	_, err := parser.ParseArgs(os.Args[1:])
	if err != nil {
		return errors.Wrap(err, "parse")
	}

	hwClock, err := rtc.SetupI2C()
	if err != nil {
		return errors.Wrap(err, "setup i2c")
	}

	switch parser.Active.Name {
	case "show":
		return show(logger, hwClock)
	case "set-system":
		return setSystemTime(logger, hwClock)
	default:
		panic("invalid command")
	}

	return nil
}

func setSystemTime(logger *zap.Logger, hwClock *ds3231.Device) error {
	t, err := hwClock.ReadTime()
	if err != nil {
		return errors.Wrap(err, "read time")
	}

	logger.Info("setting system time", zap.String("t", t.String()))
	err = util.SetTime(t)
	if err != nil {
		return errors.Wrap(err, "set time")
	}

	return nil
}

func show(logger *zap.Logger, hwClock *ds3231.Device) error {
	t, err := hwClock.ReadTime()
	if err != nil {
		return errors.Wrap(err, "read time")
	}

	fmt.Println(t.String())
	return nil
}

func main() {
	logger, _ := util2.NewLogger(false)

	if err := run(logger); err != nil {
		logger.Fatal("fatal", zap.Error(err))
	}
}
