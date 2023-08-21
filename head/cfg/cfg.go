package cfg

import (
	"github.com/minor-industries/theheads/head/motor"
	"github.com/minor-industries/theheads/head/voices"
	"time"
)

type Cfg struct {
	Instance    string
	Port        int  `envconfig:"default=8080"`
	FakeStepper bool `envconfig:"optional"`
	SensorPin   int  `envconfig:"default=21"`

	I2CBus string `envconfig:"default=1"`

	EnableMagnetSensor bool     `envconfig:"default=true"`
	MagnetSensorAddrs  []string `envconfig:"default=1f;5e"` // note semicolon to separate default values

	Motor  motor.Cfg
	Voices voices.Cfg

	Debug bool `envconfig:"optional"`

	HeartbeatInterval time.Duration `envconfig:"default=1s"`
}
