package solar

import (
	"encoding/binary"
	"github.com/cacktopus/theheads/common/standard_server"
	"github.com/goburrow/modbus"
	"github.com/jessevdk/go-flags"
	"github.com/minor-industries/platform/common/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"time"
)

func simpleGaugeFunc(name string, callback func() float64) prometheus.GaugeFunc {
	g := prometheus.NewGaugeFunc(prometheus.GaugeOpts{
		Namespace: "heads",
		Subsystem: "solar",
		Name:      name,
	}, func() float64 {
		return callback()
	})
	prometheus.MustRegister(g)
	return g
}

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

	s := SetupStats(logger)

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

	go runloop(errCh, client, s)
	go func() {
		errCh <- errors.Wrap(server.Run(), "run server")
	}()

	return errors.Wrap(<-errCh, "exit")
}

func runloop(errCh chan error, client modbus.Client, s *stats) {
	for {
		results, err := client.ReadInputRegisters(0x331A, 1)
		if err != nil {
			errCh <- errors.Wrap(err, "read input registers")
			return
		}

		s.BatteryVoltage.Set(getFloatFrom16Bit(results))
	}
}

type stats struct {
	BatteryVoltage *metrics.TimeoutGauge
}

func SetupStats(logger *zap.Logger) *stats {

	s := &stats{}

	s.BatteryVoltage = metrics.NewTimeoutGauge(time.Minute, prometheus.GaugeOpts{
		Namespace: "heads",
		Subsystem: "solar",
		Name:      "battery_voltage",
	})
	prometheus.MustRegister(s.BatteryVoltage.G)

	return s

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

func getFloatFrom16Bit(data []byte) float64 {
	return float64(binary.BigEndian.Uint16(data)) / 100
}
