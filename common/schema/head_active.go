package schema

import "time"

type Heartbeat struct {
	Component string
	Instance  string
	Start     time.Duration
	ID        string
}

func (*Heartbeat) Name() string {
	return "heartbeat"
}
