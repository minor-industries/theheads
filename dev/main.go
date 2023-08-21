package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/minor-industries/theheads/boss"
	util2 "github.com/minor-industries/theheads/boss/util"
	"github.com/minor-industries/theheads/common/discovery"
	"github.com/minor-industries/theheads/common/util"
	"github.com/minor-industries/theheads/head"
	"github.com/minor-industries/theheads/web"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

func main() {
	logger, _ := util.NewLogger(true)
	gin.SetMode(gin.ReleaseMode)

	head01 := headEnv("head-01")
	head02 := headEnv("head-02")

	logger.Info(
		"service config",
		zap.String("instance", head01.Instance),
		zap.String("addr", fmt.Sprintf("localhost:%d", head01.Port)),
	)

	var wg sync.WaitGroup
	wg.Add(1)
	done := util2.NewBroadcastCloser()

	services := &discovery.StaticDiscovery{}

	runCamera("camera-01", "pi43.raw", services)
	runCamera("camera-02", "pi42.raw", services)

	services.Register("head", head01.Instance, head01.Port)
	services.Register("head", head02.Instance, head02.Port)

	boss01 := bossEnv()
	services.Register("boss", "boss01", 8081)

	go head.Run(head01)
	go head.Run(head02)

	fakeleds01Port := util.RandomPort()
	(&fakeleds{}).Run(fakeleds01Port)
	services.Register("leds", "head-01", fakeleds01Port)

	fakeleds02Port := util.RandomPort()
	(&fakeleds{}).Run(fakeleds02Port)
	services.Register("leds", "head-02", fakeleds02Port)

	time.Sleep(50 * time.Millisecond)

	go boss.Run(boss01, services)

	if false {
		services.Register("web", "web01", 80)
		go func() {
			err := web.Run(logger, services, prometheus.NewRegistry())
			if err != nil {
				panic(err)
			}
		}()
	}

	sigc := make(chan os.Signal, 1)
	signal.Notify(sigc,
		//syscall.SIGTERM,
		syscall.SIGINT,
	)

	go func() {
		<-sigc
		wg.Done()
		done.Close()
	}()

	wg.Wait()
}

func runCamera(
	instance string,
	src string,
	services *discovery.StaticDiscovery,
) {
	camera01Port := util.RandomPort()
	services.Register("camera", instance, camera01Port)
	cmd := exec.Command("go", "run", "./cmd/camera")
	cmd.Env = append(os.Environ(),
		"DETECT_FACES=0",
		"HEIGHT=240",
		"WIDTH=320",
		fmt.Sprintf("PORT=%d", camera01Port),
		"SOURCE=file:dev/"+src,
		"INSTANCE="+instance,
	)
	go func() {
		out, err := cmd.CombinedOutput()
		if err != nil {
			panic(string(out))
		}
	}()
}
