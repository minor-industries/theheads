package util

import (
	"syscall"
	"time"
)

func SetSystemTime(time time.Time) error {
	tv := syscall.Timeval{
		Sec:  int32(time.Unix()),
		Usec: int32(time.Nanosecond() / 1e3),
	}
	return syscall.Settimeofday(&tv)
}
