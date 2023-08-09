package main

import (
	"fmt"
	"github.com/cacktopus/theheads/common/discovery"
	"github.com/cacktopus/theheads/common/util"
	"github.com/cacktopus/theheads/timesync"
	"github.com/cacktopus/theheads/timesync/cfg"
	"github.com/cacktopus/theheads/web"
	"github.com/minor-industries/codelab/logstream"
	"github.com/minor-industries/codelab/power-monitor"
	"github.com/pelletier/go-toml/v2"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"time"
)

const retryDelay = 3 * time.Second
const configHome = "/etc/env"

func run(logger *zap.Logger) error {
	args := os.Args[1:]

	if len(args) == 0 {
		return errors.New("no components given")
	}

	// TODO: ensure unique args (so we don't run multiple copies of something)

	for _, arg := range args {
		switch arg {
		case "timesync":
			env := cfg.Defaults
			if err := loadConfig(arg, &env); err != nil {
				return errors.Wrap(err, "load config")
			}
			go runComponent(logger, arg, func() error {
				timesync.Run(
					logger,
					&env,
					discovery.NewSerf("127.0.0.1:7373"),
					prometheus.NewRegistry(),
				)
				return nil
			})
		case "web":
			go runComponent(logger, arg, func() error {
				return web.Run(
					discovery.NewSerf("127.0.0.1:7373"),
					prometheus.NewRegistry(),
				)
			})
		case "power-monitor":
			go runComponent(logger, arg, func() error {
				return power_monitor.Run(
					logger,
					prometheus.NewRegistry(),
				)
			})
		case "logstream":
			go runComponent(logger, arg, func() error {
				logstream.Run(
					logger,
					prometheus.NewRegistry(),
				)
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
					err, _ := r.(error)
					// maybe best not to retry? because the process may have
					// spawned other goroutines, etc.
					logger.Error("panic", zap.Error(err))
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

func loadConfig(component string, cfg any) error {
	cfgFile := filepath.Join(configHome, component+".toml")
	content, err := os.ReadFile(cfgFile)
	if err != nil {
		return errors.Wrap(err, "readfile")
	}
	err = toml.Unmarshal(content, cfg)
	if err != nil {
		return errors.Wrap(err, "unmarshal config")
	}
	return nil
}

func main() {
	logger, _ := util.NewLogger(false)

	err := run(logger)
	logger.Fatal("run exited", zap.Error(err))
}
