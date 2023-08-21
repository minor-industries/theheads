package heartbeat

import (
	"context"
	"github.com/google/uuid"
	"github.com/minor-industries/protobuf/gen/go/heads"
	"github.com/minor-industries/theheads/common/broker"
	"github.com/minor-industries/theheads/common/schema"
	"github.com/minor-industries/theheads/head/cfg"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"sync"
	"time"
)

type beat struct {
	ID    string
	start time.Time
}

type Monitor struct {
	logger *zap.Logger
	cfg    *cfg.Cfg
	broker *broker.Broker

	lock        sync.Mutex
	currentBeat beat

	hDuration prometheus.Histogram
}

func (m *Monitor) Ack(ctx context.Context, in *heads.AckIn) (*heads.Empty, error) {
	if in.Id == "" {
		return nil, errors.New("missing id")
	}

	m.ack(in.Id)

	return &heads.Empty{}, nil
}

func NewMonitor(
	logger *zap.Logger,
	cfg *cfg.Cfg,
	b *broker.Broker,
	hDuration prometheus.Histogram,
) *Monitor {
	return &Monitor{
		logger:    logger,
		cfg:       cfg,
		broker:    b,
		hDuration: hDuration,
	}
}

func (m *Monitor) PublishLoop() {
	ticker := time.NewTicker(m.cfg.HeartbeatInterval)

	for range ticker.C {
		m.beat()
	}
}

func (m *Monitor) beat() {
	func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		if m.currentBeat.ID != "" {
			dt := time.Now().Sub(m.currentBeat.start)
			m.logger.Debug("heartbeat timed out", zap.String("id", m.currentBeat.ID))
			m.hDuration.Observe(dt.Seconds())
		}

		m.currentBeat = beat{
			ID:    uuid.New().String(),
			start: time.Now(),
		}
	}()

	msg := &schema.Heartbeat{
		Component: "head",
		Instance:  m.cfg.Instance,
		ID:        m.currentBeat.ID,
	}

	m.broker.Publish(msg)
}

func (m *Monitor) ack(id string) {
	m.lock.Lock()
	defer m.lock.Unlock()

	if id == m.currentBeat.ID {
		dt := time.Now().Sub(m.currentBeat.start)
		m.logger.Debug(
			"acked heartbeat",
			zap.String("id", id),
			zap.Float64("duration_ms", dt.Seconds()*1000.0),
		)
		m.hDuration.Observe(dt.Seconds())
		m.currentBeat = beat{}
	}
}
