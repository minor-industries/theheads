package basic

import (
	"github.com/minor-industries/theheads/boss/dj"
	"github.com/minor-industries/theheads/boss/watchdog"
	"time"
)

func Idle(sp *dj.SceneParams) {
	for {
		watchdog.Feed()

		select {
		case <-time.After(5 * time.Second):
		case <-sp.Done.Chan():
			return
		}
	}
}
