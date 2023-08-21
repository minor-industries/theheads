package jitter

import (
	"github.com/larspensjo/Go-simplex-noise/simplexnoise"
	"github.com/minor-industries/theheads/head/motor"
	"github.com/minor-industries/theheads/head/motor/seeker"
	"math/rand"
	"time"
)

type Actor struct {
	seeker *seeker.Seeker
	t0     time.Time
}

func New(numSteps int) *Actor {
	offset := time.Duration(1000*rand.Float64()) * time.Second
	return &Actor{
		seeker: seeker.New(numSteps),
		t0:     time.Now().Add(offset),
	}
}

func (a Actor) Act(pos, target int) (direction motor.Direction, done bool) {
	t := time.Now().Sub(a.t0).Seconds()

	offset := int(10 * simplexnoise.Noise1(t))
	return a.seeker.Act(pos, target+offset)
}

func (a Actor) Name() string {
	return "Jitter"
}

func (a Actor) Finish(controller *motor.Controller) {
	a.seeker.Finish(controller)
}
