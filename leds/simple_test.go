package leds

import (
	"fmt"
	"github.com/minor-industries/theheads/common/util"
	"testing"
)

func Test_randomVector(t *testing.T) {
	util.SetRandSeed()
	x, y, z := randomVector(0.75)
	fmt.Println(x, y, z)
}
