package main

import (
	"github.com/minor-industries/theheads/camera/cfg"
	"github.com/minor-industries/theheads/common/util"
)

func cameraEnv(name string, streamFile string) *cfg.Cfg {
	return &cfg.Cfg{
		BitrateKB:         400,
		CenterLine:        false,
		DetectFaces:       false, // TODO? (this was causing crashes in dev)
		DrawFrame:         "orig",
		DrawMotion:        true,
		FloodlightPin:     17,
		FOV:               64.33,
		Framerate:         25,
		Height:            240,
		Hflip:             false,
		Instance:          name,
		MotionDetectWidth: 320,
		MotionMinArea:     160,
		MotionThreshold:   16,
		Port:              util.RandomPort(),
		PrescaleWidth:     640,
		RaspiStill:        false,
		RaspividExtraArgs: nil,
		Source:            "file:" + streamFile,
		Vflip:             false,
		WarmupFrames:      0,
		Width:             320,
		WriteFacesPath:    "",
	}
}
