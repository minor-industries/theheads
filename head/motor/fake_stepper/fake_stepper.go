package fake_stepper

import (
	"github.com/minor-industries/theheads/head/motor"
)

type Motor struct {
}

func NewMotor() *Motor {
	return &Motor{}
}

func (m *Motor) Step(direction motor.Direction) error {
	return nil
}

func (m *Motor) Start() error {
	return nil
}

func (m *Motor) Off() error {
	return nil
}
