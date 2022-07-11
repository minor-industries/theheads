package boss

import (
	"context"
	"github.com/cacktopus/theheads/boss/rate_limiter"
	"github.com/cacktopus/theheads/boss/util"
	"go.uber.org/zap"
	"os"
	"time"
)

func BossRestarter(
	ctx context.Context,
	dj *DJ,
	done util.BroadcastCloser,
	logger *zap.Logger,
) {
	rate_limiter.LimitTrailing("boss.restart", time.Hour, func() {
		os.Exit(0)
	})

	done.Close()
}
