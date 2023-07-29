package schema

type Heartbeat struct {
	Component string
	Instance  string
	ID        string
}

func (*Heartbeat) Name() string {
	return "heartbeat"
}
