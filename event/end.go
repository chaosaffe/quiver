package event

import "time"

// End represents the set of lines shot before arrow retrieval
type End struct {
	Number    int
	Arrows    int
	Distance  int
	Lines     []*Line
	TimeLimit time.Duration
	LineDelay time.Duration
}
