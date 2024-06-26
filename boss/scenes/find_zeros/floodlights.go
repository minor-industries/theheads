package find_zeros

import (
	gen "github.com/minor-industries/protobuf/gen/go/heads"
	"github.com/minor-industries/theheads/boss/dj"
	"go.uber.org/zap"
)

func setupFloodLight(
	sp *dj.SceneParams,
	cameraURI string,
	state bool,
) {
	conn, err := sp.DJ.HeadManager.GetConn(cameraURI)
	if err != nil {
		sp.Logger.Error("error getting camera connection", zap.Error(err))
		return
	}

	_, err = gen.NewFloodlightClient(conn.Conn).SetState(sp.Ctx, &gen.SetStateIn{State: state})
	if err != nil {
		sp.Logger.Error("error setting floodlight state", zap.Error(err))
		return
	}
}

func setupFloodLights(sp *dj.SceneParams) {
	for _, c := range sp.DJ.Scene.CameraMap {
		cameraURI := c.URI()
		newSp := sp.WithLogger(sp.Logger.With(zap.String("camera", cameraURI)))
		go setupFloodLight(newSp, cameraURI, sp.DJ.FloodlightController())
	}
}

func setVolume(sp *dj.SceneParams) {
	for _, h := range sp.DJ.Scene.HeadMap {
		uri := h.URI()
		logger := sp.Logger.With(zap.String("uri", uri))
		go sp.DJ.HeadManager.SetVolume(
			sp.Ctx,
			logger,
			uri,
		)
	}
}
