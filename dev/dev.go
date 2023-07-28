package main

import (
	"bufio"
	cfg2 "github.com/cacktopus/theheads/boss/cfg"
	util2 "github.com/cacktopus/theheads/boss/util"
	"github.com/cacktopus/theheads/common/util"
	"github.com/cacktopus/theheads/head/cfg"
	"github.com/cacktopus/theheads/head/motor"
	"github.com/cacktopus/theheads/head/voices"
	"github.com/ory/dockertest/v3"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"io"
	"os"
	"strconv"
	"sync"
	"time"
)

func headEnv(name string) *cfg.Cfg {
	env := &cfg.Cfg{
		Instance:    name,
		Port:        util.RandomPort(),
		FakeStepper: true,
		SensorPin:   0,
		Motor: motor.Cfg{
			NumSteps:              200,
			StepSpeed:             30,
			DirectionChangePauses: 10,
		},
		Voices: voices.Cfg{
			MediaPath: os.ExpandEnv("$HOME/shared/theheads/voices"),
		},
	}
	return env
}

func bossEnv() *cfg2.Cfg {
	scenePath := os.ExpandEnv("dev/scenes/two-heads")

	boss01 := &cfg2.Cfg{
		ScenePath:            scenePath,
		SceneName:            "local-dev",
		TextSet:              "local-dev",
		SpawnPeriod:          250 * time.Millisecond,
		BossFE:               os.Getenv("BOSS_FE"),
		FloodlightController: "day-night",
		DayDetector:          []string{"time-based", "7h30m", "20h15m"},
		Debug:                true,
		CheckInTime:          500 * time.Millisecond,
		FearfulCount:         3,
		VoiceVolume:          -100,
	}
	return boss01
}

func runDockerCamera(
	logger *zap.Logger,
	done util2.BroadcastCloser,
	wg *sync.WaitGroup,
	pool *dockertest.Pool,
	instance, filename string,
) int {
	camera, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository:   "camera-arm64",
		Tag:          "latest",
		Cmd:          []string{"/bin/sleep", "60000"},
		Mounts:       []string{os.ExpandEnv("$PWD/dev") + ":/d"},
		ExposedPorts: []string{"5000"},
	})

	if err != nil {
		panic(errors.Wrap(err, "could not start resource"))
	}

	strPort := camera.GetPort("5000/tcp")
	port, err := strconv.Atoi(strPort)
	if err != nil {
		panic(err)
	}

	logger.Info("camera port", zap.Int("port", port))

	go func() {
		rOut, wOut := io.Pipe()
		rErr, wErr := io.Pipe()

		go func() {
			s := bufio.NewScanner(rOut)
			for s.Scan() {
				os.Stdout.Write(append(s.Bytes(), '\n'))
			}
		}()

		go func() {
			s := bufio.NewScanner(rErr)
			for s.Scan() {
				os.Stdout.Write(append(s.Bytes(), '\n'))
			}
		}()

		_, err := camera.Exec([]string{"/build/bin/camera"}, dockertest.ExecOptions{
			Env: []string{
				"INSTANCE=" + instance,
				"SOURCE=file:" + "/d/pi42.raw",
				"DRAW_FRAME=orig",
				"FLOODLIGHT_PIN=-1",
				"WIDTH=320",
				"HEIGHT=240",
			},
			StdOut: wOut,
			StdErr: wErr,
			TTY:    false,
		})

		if err != nil {
			panic(errors.Wrap(err, "exec"))
		}

		logger.Info("camera exited", zap.String("instance", instance))
		wg.Done()
	}()

	go func() {
		<-done.Chan()
		logger.Info("closing camera", zap.String("instance", instance))
		camera.Close()
	}()

	return port
}
