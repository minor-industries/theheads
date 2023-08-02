package main

import (
	"fmt"
	"github.com/cacktopus/theheads/common/discovery"
	"github.com/cacktopus/theheads/common/util"
	"github.com/cacktopus/theheads/timesync"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"os"
	"time"
)

func run(logger *zap.Logger) error {
	args := os.Args[1:]

	if len(args) == 0 {
		return errors.New("no components given")
	}

	// TODO: ensure unique args (so we don't run multiple copies of something)

	for _, arg := range args {
		switch arg {
		case "timesync":
			go runComponent(logger, arg, func() error {
				// TODO: want to pass logger here (with "component"), but
				// have to deal with debug flag (timesync changes it)
				timesync.Run(discovery.NewSerf("127.0.0.1:7373"))
				return nil
			})
		default:
			return fmt.Errorf("unknown component: %s", arg)
		}
	}

	select {}
}

func runComponent(
	parentLogger *zap.Logger,
	description string,
	callback func() error,
) {
	logger := parentLogger.With(zap.String("component", description))

	for {
		logger.Info("running")
		func() {
			defer func() {
				if r := recover(); r != nil {
					fmt.Printf("Panic occurred: %v. Retrying...\n", r)
					const retryDelay = 3 * time.Second
					time.Sleep(retryDelay)
				}
			}()
			err := callback()
			if err != nil {
				logger.Error("process exited", zap.Error(err))
			}
		}()
	}
}

func main() {
	logger, _ := util.NewLogger(false)

	err := run(logger)
	logger.Fatal("run exited", zap.Error(err))
}
