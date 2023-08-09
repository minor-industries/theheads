package heartbeat

import (
	"context"
	"github.com/cacktopus/theheads/common/broker"
	"github.com/cacktopus/theheads/common/util"
	"github.com/cacktopus/theheads/head/cfg"
	"github.com/minor-industries/protobuf/gen/go/heads"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func (m *Monitor) getBeat() beat {
	m.lock.Lock()
	defer m.lock.Unlock()

	return m.currentBeat
}

func TestMonitor(t *testing.T) {
	logger, _ := util.NewLogger(true)
	env := &cfg.Cfg{
		Instance: "head01",
	}
	msgBroker := broker.NewBroker()
	go msgBroker.Start()
	defer msgBroker.Stop()

	m := NewMonitor(logger, env, msgBroker)

	assert.Zero(t, m.getBeat())

	m.beat()

	b0 := m.getBeat()

	assert.NotZero(t, b0)
	assert.NotEmpty(t, b0.ID)
	assert.NotZero(t, b0.start)

	_, err := m.Ack(context.Background(), &heads.AckIn{Id: ""})
	require.Error(t, err)

	time.Sleep(5 * time.Millisecond)

	_, err = m.Ack(context.Background(), &heads.AckIn{Id: "not the real id"})
	require.NoError(t, err)

	_, err = m.Ack(context.Background(), &heads.AckIn{Id: b0.ID})
	require.NoError(t, err)

	assert.Zero(t, m.getBeat())

	t.Run("timeout", func(t *testing.T) {
		m.beat()
		b0 := m.getBeat()
		m.beat()
		b1 := m.getBeat()

		assert.NotEqual(t, b0, b1)
	})
}

func TestPublish(t *testing.T) {
	// make sure we're actually publishing

	t.Skip()
}
