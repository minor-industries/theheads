package util

import (
	"errors"
	"time"
)

func SetSystemTime(max time.Time) error {
	return errors.New("not setting time on darwin")
}
