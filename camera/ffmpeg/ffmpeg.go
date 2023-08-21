package ffmpeg

import (
	"github.com/minor-industries/theheads/camera/cfg"
	"github.com/minor-industries/theheads/common/broker"
	"github.com/minor-industries/theheads/common/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
	"gocv.io/x/gocv"
	"io"
	"sync"
	"time"
)

type Buffer struct {
	Data []byte
}

func (f *Buffer) Name() string {
	return "buffer"
}

type Ffmpeg struct {
	env    *cfg.Cfg // TODO: separate cfg
	broker *broker.Broker
	logger *zap.Logger

	once   sync.Once
	stdin  io.Writer
	stdout io.Reader

	frame          chan []byte
	gDroppedFrames prometheus.Counter
	registry       prometheus.Registerer
}

func NewFfmpeg(
	logger *zap.Logger,
	env *cfg.Cfg,
	registry prometheus.Registerer,
	broker *broker.Broker,
) *Ffmpeg {
	result := &Ffmpeg{
		logger:         logger,
		env:            env,
		registry:       registry,
		broker:         broker,
		frame:          make(chan []byte),
		gDroppedFrames: metrics.SimpleCounter(registry, "camera", "ffmpeg_dropped_frame"),
	}

	return result
}

func (ff *Ffmpeg) InitOnce(height, width int) {
	ff.once.Do(func() {
		ff.logger.Info("spawning ffmpeg", zap.Int("width", width), zap.Int("height", height))
		ff.stdin, ff.stdout = ff.spawnFfmpeg(width, height)
		go ff.feeder()
		go ff.publisher()
	})
}

func (ff *Ffmpeg) Ffmpeg(
	src *gocv.Mat,
) {
	if ff.HasWatchers() || ff.env.InitFFEarly {
		size := src.Size()
		height := size[0]
		width := size[1]

		ff.InitOnce(height, width)
	}

	matBytes := src.ToBytes()

	select {
	case ff.frame <- matBytes:
	default:
		ff.gDroppedFrames.Inc()
	}
}

func (ff *Ffmpeg) feeder() {
	for frame := range ff.frame {
		_, err := ff.stdin.Write(frame)
		if err != nil {
			panic(err) // I wonder if we ever hit this?
		}
	}
}

func (ff *Ffmpeg) publisher() {
	buf := make([]byte, 64*1024)
	for {
		nread, err := ff.stdout.Read(buf)
		if err != nil {
			time.Sleep(1 * time.Second)
			panic(err)
		}

		ff.broker.Publish(&Buffer{
			Data: buf[:nread],
		})
	}
}

func (ff *Ffmpeg) HasWatchers() bool {
	return ff.broker.SubCount() > 0
}
