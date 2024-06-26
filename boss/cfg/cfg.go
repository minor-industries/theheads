package cfg

import "time"

type Cfg struct {
	ScenePath string
	SceneName string `envconfig:"default=prod"`
	TextSet   string `envconfig:"default=prod"`

	SpawnPeriod          time.Duration `envconfig:"default=250ms"`
	BossFE               string        `envconfig:"optional"`
	FloodlightController string        `envconfig:"default=on"`

	DayDetector []string `envconfig:"default=time-based;7h30m;20h15m"`

	Debug bool `envconfig:"optional"`

	CheckInTime  time.Duration `envconfig:"default=500ms"`
	FearfulCount int           `envconfig:"default=3"`

	VoiceVolume int `envconfig:"default=-1"`
}
