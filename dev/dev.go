package main

import (
	"github.com/minor-industries/platform/common/util"
	cfg2 "github.com/minor-industries/theheads/boss/cfg"
	"github.com/minor-industries/theheads/head/cfg"
	"github.com/minor-industries/theheads/head/motor"
	"github.com/minor-industries/theheads/head/voices"
	"os"
	"time"
)

func headEnv(name string) *cfg.Cfg {
	env := &cfg.Cfg{
		Instance:           name,
		Port:               util.RandomPort(),
		FakeStepper:        true,
		SensorPin:          0,
		I2CBus:             "",
		EnableMagnetSensor: false,
		MagnetSensorAddrs:  nil,
		Motor: motor.Cfg{
			NumSteps:              200,
			StepSpeed:             30,
			DirectionChangePauses: 10,
		},
		Voices: voices.Cfg{
			MediaPath: os.ExpandEnv("$HOME/shared/theheads/voices"),
		},
		Debug:             false,
		HeartbeatInterval: time.Second,
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
