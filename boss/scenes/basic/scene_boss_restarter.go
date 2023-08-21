package basic

import (
	"github.com/minor-industries/theheads/boss/dj"
	"github.com/minor-industries/theheads/boss/rate_limiter"
	"os"
	"time"
)

func BossRestarter(sp *dj.SceneParams) {
	rate_limiter.LimitTrailing("boss.restart", time.Hour, func() {
		os.Exit(0)
	})

	sp.Done.Close()
}
