package grid

import (
	"github.com/minor-industries/platform/common/broker"
	"github.com/minor-industries/theheads/boss/scene"
	"github.com/minor-industries/theheads/common/geom"
	"github.com/minor-industries/theheads/common/schema"
	"go.uber.org/zap"
	"math"
	"sync"
)

type focalPoints struct {
	logger      *zap.Logger
	focalPoints map[string]*focalPoint
	lock        sync.Mutex
	broker      *broker.Broker
	scene       *scene.Scene
}

func (fps *focalPoints) withLock(callback func()) {
	fps.lock.Lock()
	defer fps.lock.Unlock()
	callback()
}

func (fps *focalPoints) traceFocalPoints(p0, p1 geom.Vec) bool {
	minDist := maxFloat
	var minFp *focalPoint
	var m0, m1 geom.Vec

	fps.withLock(func() {
		for _, fp := range fps.focalPoints {
			q0, q1, hit := fp.lineIntersection(p0, p1)
			if hit {
				d := q0.Sub(p0).AbsSq()
				if d < minDist {
					m0, m1 = q0, q1
					minDist = d
					minFp = fp
				}
			}
		}
	})

	// only interact with closest fp
	if minFp != nil {
		midpoint := m0.Add(m1.Sub(m0).Scale(0.5))
		to := midpoint.Sub(minFp.pos)
		minFp.pos = minFp.pos.Add(to.Scale(fps.scene.CameraSensitivity))
		minFp.refresh()
		return true
	}

	return false
}

func (fps *focalPoints) mergeOverlappingFocalPoints() {
	// this runs every update and we can tolerate some overlap, so to keep things simple,
	// if we find a single overlap just deal with it and get to any other overlaps on the
	// next run
	fps.withLock(func() {
		for _, fp0 := range fps.focalPoints {
			for id, fp1 := range fps.focalPoints {
				if fp0 == fp1 {
					continue
				}
				if fp0.overlaps(fp1, 0.5) {
					// midpoint = (fp0.pos + fp1.pos).scale(0.5
					midpoint := fp0.pos.Add(fp1.pos).Scale(0.5)
					fp0.pos = midpoint
					delete(fps.focalPoints, id)
					fps.logger.Debug(
						"focal point merged",
						zap.String("id", id),
					)
					gActiveFocalPointCount.Dec()
				}
			}
		}
	})
}

func (fps *focalPoints) publishFocalPoints() {
	var points []*schema.FocalPoint
	fps.withLock(func() {
		for _, fp := range fps.focalPoints {
			points = append(points, fp.ToMsg())
		}
	})

	msg := &schema.FocalPoints{
		FocalPoints: points,
	}

	fps.broker.Publish(msg)
}

func (fps *focalPoints) getFocalPoints() schema.FocalPoints {
	var result []*schema.FocalPoint
	fps.withLock(func() {
		for _, fp := range fps.focalPoints {
			result = append(result, fp.ToMsg())
		}
	})
	return schema.FocalPoints{FocalPoints: result}
}

func (fps *focalPoints) cleanupStale() {
	var toRemove []string
	fps.withLock(func() {
		for id, fp := range fps.focalPoints {
			if fp.isExpired(len(fps.focalPoints)) {
				toRemove = append(toRemove, id)
			}
		}

		for _, id := range toRemove {
			delete(fps.focalPoints, id)
			fps.logger.Debug(
				"focal point expired",
				zap.String("id", id),
			)
			gActiveFocalPointCount.Dec()
		}
	})

	fps.publishFocalPoints()
}
func (fps *focalPoints) closestFocalPointTo(p geom.Vec) (*schema.FocalPoint, float64) {
	minDist := maxFloat
	var minFp *schema.FocalPoint

	fps.withLock(func() {
		for _, fp := range fps.focalPoints {
			d2 := fp.pos.Sub(p).AbsSq()
			if d2 < 1e-5 {
				continue
			}
			if d2 < minDist {
				minDist = d2
				minFp = fp.ToMsg()
			}
		}
	})

	if minFp != nil {
		return minFp, math.Sqrt(minDist)
	}

	return nil, -1
}

func (fps *focalPoints) maybeSpawnFocalPoint(p geom.Vec) {
	cMaybeSpawnFocalPoint.Inc()
	newFp := NewFocalPoint(p, fpRadius, "", DefaultTTL, DefaultTTLLast)

	for _, cam := range fps.scene.CameraMap {
		fakeFp := NewFocalPoint(cam.M.Translation(), fpRadius, "", DefaultTTL, DefaultTTLLast)
		if newFp.overlaps(fakeFp, 1.0) {
			cNewFPOverlapsCamera.Inc()
			return
		}
	}

	fps.withLock(func() {
		for _, fp := range fps.focalPoints {
			if newFp.overlaps(fp, 1.0) {
				cNewFPOverlapsExisting.Inc()
				return
			}
		}

		// create new focal point
		newFp.id = assignID()
		fps.focalPoints[newFp.id] = newFp
		gActiveFocalPointCount.Inc()
		fps.logger.Debug(
			"spawning new focal point",
			zap.String("id", newFp.id),
			zap.String("pos", p.AsStr()),
		)
	})

	fps.publishFocalPoints()
}
