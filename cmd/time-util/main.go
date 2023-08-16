package main

import (
	"fmt"
	util2 "github.com/cacktopus/theheads/common/util"
	"github.com/cacktopus/theheads/timesync/rtc"
	"github.com/cacktopus/theheads/timesync/util"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"time"
)

const timeFmt = "2006-01-02 15:04:05"

type Opts struct {
	T string `short:"t" description:"time to use (in UTC). fmt: 2006-01-02 15:04:05"`

	ShowSystem struct{} `command:"show-system"`
	ShowRTC    struct{} `command:"show-rtc"`

	SetSystem struct{} `command:"set-system"`
	SetRTC    struct{} `command:"set-rtc"`
}

func (o *Opts) GetTime() (time.Time, error) {
	t, err := time.Parse(timeFmt, o.T)

	if err != nil {
		return time.Time{}, errors.Wrap(err, "parse time")
	}
	return t, nil
}

func (o *Opts) HasTime() bool {
	return o.T != ""
}

func run(logger *zap.Logger) error {
	opts := &Opts{}

	parser := flags.NewParser(opts, flags.Default)
	_, err := parser.ParseArgs(os.Args[1:])
	if err != nil {
		return errors.Wrap(err, "parse")
	}

	switch parser.Active.Name {
	case "show-system":
		return showSystem(logger, opts)
	case "show-rtc":
		return showRTC(logger)
	case "set-system":
		fmt.Println(opts)
		return setSystemTime(logger, opts)
	case "set-rtc":
		return setHWClock(logger, opts)
	default:
		panic("invalid command")
	}

	return nil
}

func showSystem(logger *zap.Logger, opts *Opts) error {
	var t time.Time
	if opts.HasTime() {
		var err error
		t, err = opts.GetTime()
		if err != nil {
			return errors.Wrap(err, "get time from options")
		}
	} else {
		t = time.Now().In(time.UTC)
	}

	logger.Info("system time", zap.String("t", t.String()))
	return nil
}

func setHWClock(logger *zap.Logger, opts *Opts) error {
	var t time.Time
	if opts.HasTime() {
		var err error
		t, err = opts.GetTime()
		if err != nil {
			return errors.Wrap(err, "get time from options")
		}
	} else {
		t = time.Now()
	}

	clock, err := rtc.SetupI2C()
	if err != nil {
		return errors.Wrap(err, "setup i2c")
	}

	t = t.In(time.UTC)

	logger.Info("setting rtc clock", zap.String("t", t.String()))
	if err := clock.SetTime(t); err != nil {
		return errors.Wrap(err, "set rtc time")
	}

	return nil
}

func setSystemTime(logger *zap.Logger, opts *Opts) error {
	var t time.Time
	fmt.Println(opts)
	if opts.HasTime() {
		var err error
		t, err = opts.GetTime()
		if err != nil {
			return errors.Wrap(err, "get time from options")
		}
	} else {
		clock, err := rtc.SetupI2C()
		if err != nil {
			return errors.Wrap(err, "setup i2c")
		}

		t, err = clock.ReadTime()
		if err != nil {
			return errors.Wrap(err, "read time")
		}
	}

	logger.Info("setting system time", zap.String("t", t.String()))
	err := util.SetSystemTime(t)
	if err != nil {
		return errors.Wrap(err, "set time")
	}

	return nil
}

func showRTC(logger *zap.Logger) error {
	hwClock, err := rtc.SetupI2C()
	if err != nil {
		return errors.Wrap(err, "setup i2c")
	}

	t, err := hwClock.ReadTime()
	if err != nil {
		return errors.Wrap(err, "read time")
	}

	logger.Info("time", zap.String("t", t.String()))
	return nil
}

func main() {
	logger, _ := util2.NewLogger(false)

	if err := run(logger); err != nil {
		logger.Fatal("fatal", zap.Error(err))
	}
}
