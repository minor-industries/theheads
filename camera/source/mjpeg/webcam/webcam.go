package webcam

import (
	"fmt"
	"github.com/minor-industries/theheads/camera/cfg"
	"github.com/minor-industries/theheads/camera/recorder"
	"github.com/minor-industries/theheads/camera/recorder/simple_recorder"
	"github.com/minor-industries/theheads/camera/source/mjpeg/lib"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
	"os/exec"
	"strings"
)

type MjpegWebcam struct {
	frames chan *lib.Frame
	*simple_recorder.Recorder
}

func (m *MjpegWebcam) Info() recorder.RecorderInfo {
	return recorder.RecorderInfo{FileExtension: "mjpeg"}
}

func NewMjpegWebcam(logger *zap.Logger, env *cfg.Cfg) (*MjpegWebcam, error) {
	args := []string{
		"ffmpeg",
		"-f", "v4l2",
		"-framerate", fmt.Sprintf("%d", env.Framerate),
		"-video_size", fmt.Sprintf("%dx%d", env.Width, env.Height),
		"-input_format", "mjpeg",
		"-i", "/dev/video0",
		"-c", "copy",
		"-f", "mjpeg",
		"-",
	}

	logger.Info("running input ffmpeg", zap.String("cmd", strings.Join(args, " ")))

	cmd := exec.Command(args[0], args[1:]...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, errors.Wrap(err, "stdout pipe")
	}

	frames := make(chan *lib.Frame)
	recorder := simple_recorder.New(env.RecorderBufsize)

	go lib.DecodeMjpeg(env, stdout, func(frame *lib.Frame) {
		recorder.AddFrame(frame)
		frames <- frame
	})

	err = cmd.Start()
	if err != nil {
		if err != nil {
			return nil, errors.Wrap(err, "start")
		}
	}

	return &MjpegWebcam{
		frames:   frames,
		Recorder: recorder,
	}, nil
}

func (m *MjpegWebcam) Grab(dst *gocv.Mat) error {
	frame := <-m.frames
	rgb, err := gocv.ImageToMatRGB(frame.Image)
	if err != nil {
		panic(err)
	}
	defer rgb.Close()
	gocv.CvtColor(rgb, dst, gocv.ColorBGRToGray)
	return nil
}
