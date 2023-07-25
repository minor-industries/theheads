package solar

import (
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"github.com/cacktopus/theheads/common/standard_server"
	"github.com/goburrow/modbus"
	"github.com/jessevdk/go-flags"
	"github.com/minor-industries/platform/common/metrics"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"math"
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

	metrics_ := []*metric{
		{Name: "array_voltage", CB: float64by100(0x3100)},
		{Name: "array_current", CB: float64by100(0x3101)},

		{Name: "battery_voltage", CB: float64by100(0x331A)},
		{Name: "battery_current", CB: signedInt32by100(0x331B)},
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
			val, err := m.CB(client)
			if err != nil {
				logger.Error("error reading modbus", zap.Error(err))
				continue
			}

			m.G.Set(val)
		}
	}
}

type converter func(client modbus.Client) (float64, error)

type metric struct {
	Name string
	CB   converter
	G    *metrics.TimeoutGauge
}

func float64by100(addr uint16) converter {
	return func(client modbus.Client) (float64, error) {
		data, err := client.ReadInputRegisters(addr, 1)
		if err != nil {
			return 0, errors.Wrap(err, "read input registers")
		}
		return getFloatFrom16Bit(data), nil
	}
}

func signedInt32by100(addr uint16) converter {
	return func(client modbus.Client) (float64, error) {
		data, err := client.ReadInputRegisters(addr, 2)
		if err != nil {
			return math.NaN(), errors.Wrap(err, "read input registers")
		}

		if len(data) != 4 {
			return math.NaN(), fmt.Errorf("invalid data length")
		}

		fmt.Println(hex.Dump(data))

		val := getFloatFromSigned32Bit(data)
		return val, nil
	}
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

func getFloatFrom16Bit(data []byte) float64 {
	return float64(binary.BigEndian.Uint16(data)) / 100
}

func getFloatFromSigned32Bit(data []byte) float64 {
	return float64(getSigned32BitData(data)) / 100
}

func getUnsigned32BitData(data []byte) uint32 {
	var buf []byte
	buf = append(buf, data[2:4]...)
	buf = append(buf, data[0:2]...)
	return binary.BigEndian.Uint32(buf)
}

func getSigned32BitData(data []byte) int32 {
	var buf []byte
	buf = append(buf, data[2:4]...)
	buf = append(buf, data[0:2]...)
	return int32(binary.BigEndian.Uint32(buf))
}
