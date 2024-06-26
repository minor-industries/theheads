package file

import (
	"github.com/minor-industries/theheads/camera/cfg"
	"github.com/minor-industries/theheads/camera/recorder"
	"github.com/minor-industries/theheads/camera/recorder/simple_recorder"
	"github.com/minor-industries/theheads/camera/source/mjpeg/lib"
	"gocv.io/x/gocv"
	"os"
	"time"
)

type MjpegFileStreamer struct {
	frames chan *lib.Frame
	*simple_recorder.Recorder
}

func (m *MjpegFileStreamer) Info() recorder.RecorderInfo {
	return recorder.RecorderInfo{FileExtension: "mjpeg"}
}

type circularReader struct {
	circularBuf []byte
	pos         int
}

func (c *circularReader) Read(p []byte) (int, error) {
	sz := len(p)

	for len(p) > 0 {
		chunk := c.circularBuf[c.pos:]
		n := copy(p, chunk)

		c.pos += n
		if c.pos == len(c.circularBuf) {
			c.pos = 0
		}

		p = p[n:]
	}

	return sz, nil
}

func NewMjpegFileStreamer(env *cfg.Cfg) *MjpegFileStreamer {
	content, err := os.ReadFile(os.ExpandEnv("$HOME/14_32_31.mjpeg"))
	if err != nil {
		panic(err)
	}

	frameBuf := make(chan *lib.Frame)
	inputFrames := make(chan *lib.Frame)
	simpleRec := simple_recorder.New(env.RecorderBufsize)

	go lib.DecodeMjpeg(env, &circularReader{circularBuf: content}, func(frame *lib.Frame) {
		simpleRec.AddFrame(frame)
		frameBuf <- frame
	})

	go func() {
		ticker := time.NewTicker(time.Second / 30)
		for range ticker.C {
			frame := <-frameBuf
			if err := lib.FrameIsValid(frame.Raw); err != nil {
				panic(err)
			}
			inputFrames <- frame
		}
	}()

	return &MjpegFileStreamer{
		frames:   inputFrames,
		Recorder: simpleRec,
	}
}

func (m *MjpegFileStreamer) Grab(dst *gocv.Mat) error {
	frame := <-m.frames
	rgb, err := gocv.ImageToMatRGB(frame.Image)
	if err != nil {
		panic(err)
	}
	defer rgb.Close()
	gocv.CvtColor(rgb, dst, gocv.ColorBGRToGray)
	return nil
}
