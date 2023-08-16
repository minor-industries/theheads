package main

import (
	"fmt"
	"github.com/cacktopus/theheads/timesync/rtc"
	"github.com/cacktopus/theheads/timesync/util"
	"github.com/jessevdk/go-flags"
	"github.com/pkg/errors"
	"os"
	"time"
)

const timeFmt = "2006-01-02 15:04:05 MST"

type Opts struct {
	T string `short:"t" description:"time to use (in UTC). fmt: 2006-01-02 15:04:05 MST"`

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

	if t.Location() != time.UTC {
		return time.Time{}, errors.New("time must be specified in UTC")
	}

	return t, nil
}

func (o *Opts) HasTime() bool {
	return o.T != ""
}

func run() error {
	opts := &Opts{}

	parser := flags.NewParser(opts, flags.Default)
	_, err := parser.ParseArgs(os.Args[1:])
	if err != nil {
		return errors.Wrap(err, "parse")
	}

	switch parser.Active.Name {
	case "show-system":
		return showSystem(opts)
	case "show-rtc":
		return showRTC()
	case "set-system":
		return setSystemTime(opts)
	case "set-rtc":
		return setHWClock(opts)
	default:
		panic("invalid command")
	}

	return nil
}

func showSystem(opts *Opts) error {
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

	fmt.Println("system time:", t.Format(timeFmt))
	return nil
}

func setHWClock(opts *Opts) error {
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

	fmt.Println("setting rtc clock:", t.Format(timeFmt))
	if err := clock.SetTime(t); err != nil {
		return errors.Wrap(err, "set rtc time")
	}

	return nil
}

func setSystemTime(opts *Opts) error {
	var t time.Time
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

	fmt.Println("setting system time:", t.Format(timeFmt))
	err := util.SetSystemTime(t)
	if err != nil {
		return errors.Wrap(err, "set time")
	}

	return nil
}

func showRTC() error {
	hwClock, err := rtc.SetupI2C()
	if err != nil {
		return errors.Wrap(err, "setup i2c")
	}

	t, err := hwClock.ReadTime()
	if err != nil {
		return errors.Wrap(err, "read time")
	}

	fmt.Println("rtc time:", t.Format(timeFmt))
	return nil
}

func main() {
	if err := run(); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "error: %s\n", err.Error())
		os.Exit(1)
	}
}
