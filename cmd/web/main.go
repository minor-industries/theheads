package main

import (
	"github.com/minor-industries/platform/common/discovery"
	"github.com/minor-industries/platform/common/metrics"
	"github.com/minor-industries/platform/common/util"
	"github.com/minor-industries/theheads/web"
)

func main() {
	logger, _ := util.NewLogger(false)

	if err := web.Run(
		logger,
		discovery.NewSerf("127.0.0.1:7373"),
		metrics.NewRegistry(),
	); err != nil {
		panic(err)
	}
}
