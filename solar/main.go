package solar

import (
	"fmt"
	"github.com/cacktopus/theheads/common/standard_server"
	"github.com/cacktopus/theheads/solar/convert"
	"github.com/goburrow/modbus"
	"github.com/jessevdk/go-flags"
	"github.com/minor-industries/platform/common/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"time"
)

var opt struct {
	SerialPort string `long:"serial-port" env:"SERIAL_PORT" default:"/dev/ttyXRUSB0"`
	Port       int    `long:"port" env:"PORT" default:"8089"`
}

func Run() error {
	logger, _ := zap.NewProduction()

	_, err := flags.Parse(&opt)
	if err != nil {
		return errors.Wrap(err, "parse")
	}

	handler := modbus.NewRTUClientHandler(opt.SerialPort)
	handler.BaudRate = 115200
	handler.DataBits = 8
	handler.Parity = "N"
	handler.StopBits = 1
	handler.SlaveId = 1
	handler.Timeout = 5 * time.Second

	err = handler.Connect()
	if err != nil {
		return errors.Wrap(err, "connect to serial port")
	}

	client := modbus.NewClient(handler)

	const d2 = 1.0 / 100.0

	metrics_ := []*metric{
		{Name: "array_voltage", CB: convert.UnsignedInt16(0x3100, d2)},
		{Name: "array_current", CB: convert.UnsignedInt16(0x3101, d2)},
		{Name: "array_power", CB: convert.UnsignedInt32(0x3102, d2)},

		{Name: "load_voltage", CB: convert.UnsignedInt16(0x310C, d2)},
		{Name: "load_current", CB: convert.UnsignedInt16(0x310D, d2)},
		{Name: "load_power", CB: convert.UnsignedInt32(0x310E, d2)},

		//{Name: "", CB: nil},
		//{Name: "", CB: nil},
		//{Name: "", CB: nil},
		//{Name: "", CB: nil},
		//{Name: "", CB: nil},

		{Name: "battery_voltage", CB: convert.UnsignedInt16(0x331A, d2)},
		{Name: "battery_current", CB: convert.SignedInt32(0x331B, d2)},
		{Name: "battery_power", CB: convert.UnsignedInt32(0x3106, d2)},
		{Name: "battery_state_of_charge", CB: convert.UnsignedInt16(0x311A, 1.0)},

		{Name: "battery_temperature_celsius", CB: convert.SignedInt16(0x3110, d2)},
	}

	SetupMetrics(metrics_)

	server, err := standard_server.NewServer(&standard_server.Config{
		Logger:    logger,
		Port:      opt.Port,
		GrpcSetup: nil,
		HttpSetup: nil,
	})
	if err != nil {
		return errors.Wrap(err, "new server")
	}

	errCh := make(chan error)

	go runloop(logger, client, metrics_)
	go func() {
		errCh <- errors.Wrap(server.Run(), "run server")
	}()

	return errors.Wrap(<-errCh, "exit")
}

func runloop(logger *zap.Logger, client modbus.Client, metrics_ []*metric) {
	ticker := time.NewTicker(5 * time.Second)

	for range ticker.C {
		for _, m := range metrics_ {
			fmt.Println("\n" + m.Name)
			val, err := m.CB(client)
			if err != nil {
				logger.Error("error reading modbus", zap.Error(err))
				continue
			}

			m.G.Set(val)
		}
	}
}

type metric struct {
	Name string
	CB   convert.Converter
	G    *metrics.TimeoutGauge
}

func SetupMetrics(metrics_ []*metric) {
	for _, m := range metrics_ {
		m.G = metrics.NewTimeoutGauge(time.Minute, prometheus.GaugeOpts{
			Namespace: "heads",
			Subsystem: "solar",
			Name:      m.Name,
		})
		prometheus.MustRegister(m.G.G)
	}

	//newStat("array_voltage", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.ArrayVoltage)
	//})
	//
	//newStat("array_power", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.ArrayPower)
	//})
	//
	//newStat("array_current", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.ArrayCurrent)
	//})
	//
	//newStat("battery_voltage", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.BatteryVoltage)
	//})
	//
	//newStat("battery_current", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.BatteryCurrent)
	//})
	//
	//newStat("battery_state_of_charge", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.BatterySOC)
	//})
	//
	//newStat("battery_temperature_celsius", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.BatteryTemp)
	//})
	//
	//newStat("battery_temperature_fahrenheit", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.BatteryTemp)*9.0/5.0 + 32.0
	//})
	//
	//newStat("load_voltage", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.LoadVoltage)
	//})
	//
	//newStat("load_current", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.LoadCurrent)
	//})
	//
	//newStat("load_power", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.LoadPower)
	//})
	//
	//newStat("device_temperature_celsius", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.BatteryTemp)
	//})
	//
	//newStat("device_temperature_fahrenheit", func(status *gotracer.TracerStatus) float64 {
	//	return float64(status.BatteryTemp)*9.0/5.0 + 32.0
	//})
}
