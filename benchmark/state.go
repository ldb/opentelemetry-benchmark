package benchmark

type State int

const (
	Unknown State = iota
	Uninitialized
	Configured
	Running
	Finished
	Stopped
)

var stateNames = [...]string{
	"unknown",
	"uninitialized",
	"configured",
	"running",
	"finished",
	"stopped",
}

func (s State) String() string {
	if int(s) >= len(stateNames) {
		return stateNames[0]
	}
	return stateNames[s]
}

func StateFrom(s string) State {
	for k, name := range stateNames {
		if name == s {
			return State(k)
		}
	}
	return Unknown
}
