package web

import (
	"github.com/minor-industries/platform/common/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"os/exec"
	"strconv"
	"strings"
	"time"
)

/*
https://raspberrypi.stackexchange.com/questions/60593/how-raspbian-detects-under-voltage

Bit Hex value   Meaning
0          1    Under-voltage detected
1          2    Arm frequency capped
2          4    Currently throttled
3          8    Soft temperature limit active
16     10000    Under-voltage has occurred
17     20000    Arm frequency capping has occurred
18     40000    Throttling has occurred
19     80000    Soft temperature limit has occurred
*/

type metric struct {
	name string
	mask int64
	g    *metrics.TimeoutGauge
}

func monitorLowVoltage(
	logger *zap.Logger,
	registry *prometheus.Registry,
	errCh chan error,
) {
	ticker := time.NewTicker(5 * time.Second)

	allMetrics := []*metric{
		{name: "low_voltage_now", mask: 0x1},
		{name: "frequency_capped_now", mask: 0x2},
		{name: "throttled_now", mask: 0x4},
		{name: "soft_temperature_limit_now", mask: 0x8},
		{name: "low_voltage_observed", mask: 0x10000},
		{name: "frequency_capped_observed", mask: 0x20000},
		{name: "throttled_observed", mask: 0x40000},
		{name: "soft_temperature_limit_observed", mask: 0x80000},
	}

	for _, m := range allMetrics {
		m.g = metrics.NewTimeoutGauge(time.Minute, prometheus.GaugeOpts{
			Namespace: "rpi",
			Subsystem: "power",
			Name:      m.name,
		})
		registry.MustRegister(m.g.G)
	}

	for range ticker.C {
		cmd := exec.Command("vcgencmd", "get_throttled")
		buf, err := cmd.CombinedOutput()
		output := string(buf)
		if err != nil {
			logger.Error(
				"vcgencmd error",
				zap.Error(err),
				zap.String("output", output),
			)
			continue
		}

		parts := strings.Split(strings.TrimSpace(output), "=")
		if len(parts) != 2 {
			logger.Error("vcgencmd invalid output", zap.String("output", output))
			continue
		}

		val, err := strconv.ParseInt(parts[1][2:], 16, 64)
		if err != nil {
			logger.Error("vcgencmd parse error", zap.Error(err))
			continue
		}

		for _, m := range allMetrics {
			if val&m.mask != 0 {
				m.g.Set(1.0)
			} else {
				m.g.Set(0.0)
			}
		}
	}
}
