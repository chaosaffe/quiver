package event

import "time"

// represents the full competition

type Round struct {
	Name     string
	Ends     []*End
	EndDelay time.Duration
}
