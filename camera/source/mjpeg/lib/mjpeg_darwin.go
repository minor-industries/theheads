package lib

import (
	"github.com/minor-industries/theheads/camera/cfg"
	"image"
	"io"
)

var StartFrameMarker = []byte{0xff, 0xd8}
var EndFrameMarker = []byte{0xff, 0xd9}

type Frame struct {
	Raw   []byte
	Image image.Image
}

func (f *Frame) Write(first bool, writer io.Writer) error {
	panic("not implemented")
}

func (f *Frame) IsValid() error {
	panic("not implemented")
}

func FrameIsValid(frame []byte) error {
	panic("not implemented")
}

func DecodeMjpeg(
	env *cfg.Cfg,
	input io.Reader,
	callback func(*Frame),
) {
	panic("not implemented")
}
