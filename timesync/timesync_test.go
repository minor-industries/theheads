package timesync

import (
	"github.com/minor-industries/theheads/common/discovery"
	"github.com/minor-industries/theheads/common/util"
	"github.com/minor-industries/theheads/timesync/cfg"
	"testing"
	"time"
)

func Test_sync(t *testing.T) {
	logger, _ := util.NewLogger(false)
	services := &discovery.StaticDiscovery{}

	c1 := &cfg.Config{
		Port:       util.RandomPort(),
		RTC:        true,
		Interval:   5 * time.Second,
		MinSources: 1,
	}
	c2 := &cfg.Config{
		Port:       util.RandomPort(),
		RTC:        true,
		Interval:   6 * time.Second,
		MinSources: 1,
	}
	c3 := &cfg.Config{
		Port:       util.RandomPort(),
		RTC:        false,
		Interval:   7 * time.Second,
		MinSources: 1,
	}

	go run(logger, c1, services, nil)
	go run(logger, c2, services, nil)
	go run(logger, c3, services, nil)

	time.Sleep(200 * time.Millisecond)

	services.Register("timesync", "timesync-01", c1.Port)
	services.Register("timesync", "timesync-02", c2.Port)
	services.Register("timesync", "timesync-03", c3.Port)

	select {}
}
