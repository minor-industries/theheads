package cfg

import "time"

type Config struct {
	Port       int           `envconfig:"default=8086" toml:",omitempty"`
	RTC        bool          `envconfig:"optional" toml:",omitempty"`
	Interval   time.Duration `envconfig:"default=60s" toml:",omitempty"`
	MinSources int           `envconfig:"default=2" toml:",omitempty"`
}

var Defaults = Config{
	Port:       8086,
	RTC:        false,
	Interval:   60 * time.Second,
	MinSources: 2,
}
