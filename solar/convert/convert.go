package convert

import (
	"encoding/binary"
	"fmt"
	"github.com/goburrow/modbus"
	"github.com/pkg/errors"
	"math"
)

type Converter func(client modbus.Client) (float64, error)

func UnsignedInt16(addr uint16, scale float64) Converter {
	return func(client modbus.Client) (float64, error) {
		data, err := client.ReadInputRegisters(addr, 1)
		if err != nil {
			return 0, errors.Wrap(err, "read input registers")
		}

		if len(data) != 2 {
			return math.NaN(), fmt.Errorf("invalid data length")
		}

		val := unsigned16Bit(data) * scale
		return val, nil
	}
}

func SignedInt16(addr uint16, scale float64) Converter {
	return func(client modbus.Client) (float64, error) {
		data, err := client.ReadInputRegisters(addr, 1)
		if err != nil {
			return 0, errors.Wrap(err, "read input registers")
		}

		if len(data) != 2 {
			return math.NaN(), fmt.Errorf("invalid data length")
		}

		val := signed16Bit(data) * scale
		return val, nil
	}
}

func UnsignedInt32(addr uint16, scale float64) Converter {
	return func(client modbus.Client) (float64, error) {
		data, err := client.ReadInputRegisters(addr, 2)
		if err != nil {
			return math.NaN(), errors.Wrap(err, "read input registers")
		}

		if len(data) != 4 {
			return math.NaN(), fmt.Errorf("invalid data length")
		}

		val := unsigned32Bit(data) * scale

		return val, nil
	}
}

func SignedInt32(addr uint16, scale float64) Converter {
	return func(client modbus.Client) (float64, error) {
		data, err := client.ReadInputRegisters(addr, 2)
		if err != nil {
			return math.NaN(), errors.Wrap(err, "read input registers")
		}

		if len(data) != 4 {
			return math.NaN(), fmt.Errorf("invalid data length")
		}

		val := signed32Bit(data) * scale
		return val, nil
	}
}

func unsigned16Bit(data []byte) float64 {
	return float64(binary.BigEndian.Uint16(data))
}

func signed16Bit(data []byte) float64 {
	return float64(int16(binary.BigEndian.Uint16(data)))
}

func signed32Bit(data []byte) float64 {
	return float64(asInt32(data))
}

func unsigned32Bit(data []byte) float64 {
	return float64(asUint32(data))
}

func asUint32(data []byte) uint32 {
	var buf []byte
	buf = append(buf, data[2:4]...)
	buf = append(buf, data[0:2]...)
	return binary.BigEndian.Uint32(buf)
}

func asInt32(data []byte) int32 {
	return int32(asUint32(data))
}
