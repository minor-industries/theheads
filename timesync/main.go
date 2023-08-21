package timesync

import (
	"github.com/coreos/go-systemd/daemon"
	gen "github.com/minor-industries/protobuf/gen/go/heads"
	"github.com/minor-industries/theheads/common/discovery"
	"github.com/minor-industries/theheads/common/retry"
	"github.com/minor-industries/theheads/common/standard_server"
	"github.com/minor-industries/theheads/timesync/cfg"
	"github.com/minor-industries/theheads/timesync/rtc"
	"github.com/minor-industries/theheads/timesync/rtc/ds3231"
	"github.com/minor-industries/theheads/timesync/server"
	"github.com/minor-industries/theheads/timesync/sync"
	"github.com/minor-industries/theheads/timesync/util"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

func run(
	logger *zap.Logger,
	env *cfg.Config,
	discovery discovery.Discovery,
	registry *prometheus.Registry,
) {
	setupMetrics(registry)

	if env.MinSources < 1 {
		panic("MIN_SOURCES must be greater than zero")
	}

	go func() {
		err := retry.Retry(5, 2*time.Second, func(attempt int) error {
			err := sync.Synctime(env, logger, discovery, true)
			if err != nil {
				logger.Debug("fast sync error", zap.Int("attempt", attempt), zap.Error(err))
			}
			return err
		})
		if err != nil {
			logger.Error("fast sync failed", zap.Error(err))
		} else {
			// raise the logging level if we get a successful sync
			logger = logger.WithOptions(zap.IncreaseLevel(zap.InfoLevel))
		}
		logger.Info("notifying systemd")
		_, err = daemon.SdNotify(true, daemon.SdNotifyReady)
		if err != nil {
			logger.Warn("systemd notify failed", zap.Error(err))
		}
		for range time.NewTicker(env.Interval).C {
			err = sync.Synctime(env, logger, discovery, false)
			if err != nil {
				logger.Debug("sync error")
			} else {
				// raise the logging level if we get a successful sync
				logger = logger.WithOptions(zap.IncreaseLevel(zap.InfoLevel))
			}
		}
	}()

	h := &server.Handler{
		RTC: env.RTC,
	}

	server, err := standard_server.NewServer(&standard_server.Config{
		Logger: logger,
		Port:   env.Port,
		GrpcSetup: func(grpcServer *grpc.Server) error {
			gen.RegisterTimeServer(grpcServer, h)
			return nil
		},
		Registry: registry,
	})
	if err != nil {
		panic(err)
	}

	err = server.Run()
	if err != nil {
		panic(err)
	}
}

func setupMetrics(registry *prometheus.Registry) {
	registry.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "heads",
		Subsystem: "timesync",
		Name:      "current_timestamp",
	}, util.Now))
}

func readRTCTime(rtc *ds3231.Device, logger *zap.Logger) bool {
	if rtc == nil {
		logger.Debug("no rtc available")
		return false
	}

	t, err := rtc.ReadTime()
	if err != nil {
		logger.Debug("error reading rtc time", zap.Error(err))
		return false
	}

	now := time.Now().In(time.UTC)

	dt := now.Sub(t)
	if dt < 0 {
		dt = -dt
	}

	correct := dt < 5*time.Second

	logger.Debug(
		"read rtc time",
		zap.String("rtc_time", t.UTC().String()),
		zap.String("system_time", now.String()),
		zap.Bool("correct", correct),
	)

	return correct
}

func Run(
	logger *zap.Logger,
	env *cfg.Config,
	discover discovery.Discovery,
	registry *prometheus.Registry,
) {
	rtClock, err := rtc.SetupI2C()
	if err != nil {
		logger.Warn("error setting up i2c", zap.Error(err))
	}

	correct := readRTCTime(rtClock, logger)
	if correct {
		logger = logger.WithOptions(zap.IncreaseLevel(zap.InfoLevel))
	}

	run(logger, env, discover, registry)
}
