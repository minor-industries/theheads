package find_zeros

import (
	"github.com/minor-industries/theheads/boss/dj"
	"github.com/minor-industries/theheads/boss/scenes"
	"github.com/minor-industries/theheads/boss/watchdog"
)

func FindZeros(sp *dj.SceneParams) {
	sp.DJ.Scene.ClearFearful()

	sp.DJ.HeadManager.CheckIn(
		sp.Ctx,
		sp.Logger,
		sp.DJ.Scene,
		sp.DJ.Boss.Env.CheckInTime,
	)

	scenes.SceneSetup(sp, "rainbow")

	go setupFloodLights(sp)
	go setVolume(sp)
	go findHeadZeros(sp)

	<-sp.Done.Chan()
	sp.Logger.Info("Exiting FindZeros")
	watchdog.Feed()
}
