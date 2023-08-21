package app

import (
	"github.com/minor-industries/platform/common/broker"
	"github.com/minor-industries/platform/common/standard_server"
	"github.com/minor-industries/platform/schema"
	"github.com/minor-industries/theheads/boss/cfg"
	"github.com/minor-industries/theheads/boss/day"
	"github.com/minor-industries/theheads/boss/grid"
	"github.com/minor-industries/theheads/boss/head_manager"
	"github.com/minor-industries/theheads/boss/scene"
	"github.com/minor-industries/theheads/boss/services"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"io/fs"
)

type Boss struct {
	Logger      *zap.Logger
	Env         *cfg.Cfg
	Broker      *broker.Broker
	Grid        *grid.Grid
	Directory   *services.Directory
	Server      *standard_server.Server
	Scene       *scene.Scene
	Frontend    fs.FS
	DayDetector day.Detector
	HeadManager *head_manager.HeadManager
}

func (b *Boss) SetupMetrics() {
	prometheus.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "heads",
		Subsystem: "boss",
		Name:      "is_day",
	}, func() float64 {
		if b.DayDetector.IsDay() {
			return 1.0
		}
		return 0.0
	}))

	prometheus.MustRegister(prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "heads",
		Subsystem: "boss",
		Name:      "fearful_heads",
	}, func() float64 {
		result := 0.0
		for _, head := range b.Scene.HeadMap {
			if head.Fearful() {
				result += 1.0
			}
		}
		return result
	}))
}

func (b *Boss) processHeartbeat(msg *schema.Heartbeat) {
	logger := b.Logger.With(zap.String("instance", msg.Instance))
	switch msg.Component {
	case "head":
		head, ok := b.Scene.HeadMap[msg.Instance]
		if !ok {
			logger.Warn("heartbeat: unknown instance")
			return
		}
		if err := b.HeadManager.AckHeartbeat(head.URI(), msg.ID); err != nil {
			logger.Warn("failed to ack heartbeat", zap.Error(err))
			return
		}
	}
}
