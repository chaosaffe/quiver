package timer

import (
	"time"
)

type Timer struct {
	Duration    time.Duration
	Resolution  time.Duration
	tickChannel chan time.Duration
	stopChannel chan bool
	running     bool
}

func NewTimer(d, r time.Duration) *Timer {
	return &Timer{
		Duration:    d,
		Resolution:  r,
		tickChannel: make(chan time.Duration),
		running:     false,
	}
}

func (t *Timer) Start() {
	if !t.running {
		t.running = true
		go t.run()
	}
}

func (t *Timer) run() {
	for t.Duration >= 0 {
		select {
		case <-t.stopChannel:
			t.running = false
			return
		default:
			t.tickChannel <- t.Duration
			t.Duration -= t.Resolution
			time.Sleep(t.Resolution)
		}
	}
}

func (t *Timer) Stop() {
	if t.running {
		t.stopChannel <- true
	}
}

func (t *Timer) TickChannel() chan time.Duration {
	return t.tickChannel
}
