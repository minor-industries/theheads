package heartbeat

import (
	"context"
	"github.com/cacktopus/theheads/common/broker"
	"github.com/cacktopus/theheads/common/schema"
	"github.com/cacktopus/theheads/gen/go/heads"
	"github.com/cacktopus/theheads/head/cfg"
	"github.com/google/uuid"
	"github.com/pkg/errors"
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
}

func (m *Monitor) Ack(ctx context.Context, in *heads.AckIn) (*heads.Empty, error) {
	if in.Id == "" {
		return nil, errors.New("missing id")
	}

	m.ack(in.Id)

	return &heads.Empty{}, nil
}

func NewMonitor(logger *zap.Logger, cfg *cfg.Cfg, b *broker.Broker) *Monitor {
	return &Monitor{
		logger: logger,
		cfg:    cfg,
		broker: b,
	}
}

func (m *Monitor) PublishLoop() {
	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {
		m.beat()
	}
}

func (m *Monitor) beat() {
	func() {
		m.lock.Lock()
		defer m.lock.Unlock()

		if m.currentBeat.ID != "" {
			m.logger.Info("heartbeat timed out")
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
		m.logger.Info(
			"acked heartbeat",
			zap.String("id", id),
			zap.Int64("duration_ms", dt.Milliseconds()),
		)
		m.currentBeat = beat{}
	}
}
