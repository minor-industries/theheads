package heartbeat

import (
	"context"
	"github.com/cacktopus/theheads/common/broker"
	"github.com/cacktopus/theheads/common/schema"
	"github.com/cacktopus/theheads/gen/go/heads"
	"github.com/cacktopus/theheads/head/cfg"
	"github.com/google/uuid"
	"github.com/pkg/errors"
	"time"
)

type Monitor struct {
	cfg *cfg.Cfg
	b   *broker.Broker
	t0  time.Time
}

func (m *Monitor) Ack(ctx context.Context, in *heads.AckIn) (*heads.Empty, error) {
	return nil, errors.New("not implemented yet")
}

func NewMonitor(cfg *cfg.Cfg, b *broker.Broker) *Monitor {
	return &Monitor{
		b:   b,
		cfg: cfg,
		t0:  time.Now(),
	}
}

func (m *Monitor) PublishLoop() {
	for {
		time.Sleep(5 * time.Second)

		id := uuid.New().String()

		msg := &schema.Heartbeat{
			Component: "head",
			Instance:  m.cfg.Instance,
			Start:     time.Now().Sub(m.t0),
			ID:        id,
		}

		m.b.Publish(msg)
	}
}
