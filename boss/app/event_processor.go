package app

import (
	geom2 "github.com/minor-industries/platform/common/geom"
	"github.com/minor-industries/platform/schema"
	"github.com/minor-industries/theheads/boss/rate_limiter"
	"go.uber.org/zap"
	"time"
)

func (b *Boss) ProcessEvents() {
	msgs := b.Broker.Subscribe()

	for i := range msgs {
		switch msg := i.(type) {
		case *schema.MotionDetected:
			b.processMotion(msg)
		case *schema.FaceDetected:
			b.processFaceDetected(msg)
		case *schema.Heartbeat:
			b.processHeartbeat(msg)
		}
	}
}

func (b *Boss) processMotion(msg *schema.MotionDetected) {
	cam, ok := b.Scene.CameraMap[msg.CameraName]
	if !ok {
		rate_limiter.Debounce(
			"detected motion from unknown camera: "+msg.CameraName,
			time.Minute,
			func() {
				b.Logger.Error("detected motion from unknown camera", zap.String("camera", msg.CameraName))
			},
		)
		return
	}

	rotz := geom2.Rotz(msg.Position)

	p0 := geom2.ZeroVec()
	p1 := rotz.MulVec(geom2.NewVec(10, 0))

	m := cam.Stand.M.Mul(cam.M)

	p0 = m.MulVec(p0)
	p1 = m.MulVec(p1)

	b.Grid.Trace(msg.CameraName, p0, p1)
}

func (b *Boss) processFaceDetected(msg *schema.FaceDetected) {
	err := b.Scene.OnFaceDetected(msg)
	if err != nil {
		b.Logger.Error("error processing face-detected", zap.Error(err))
	}
}
