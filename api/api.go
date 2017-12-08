package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/chaosaffe/quiver/timer"
	"github.com/gorilla/mux"
)

const eventWrapper = "data: %s\n\n"

var hub = &Connections{
	Clients:      make(map[chan TimerEvent]bool),
	AddClient:    make(chan (chan TimerEvent)),
	RemoveClient: make(chan (chan TimerEvent)),
	Messages:     make(chan TimerEvent),
}

func setStreamHeaders(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
}

func timerEventsHandler(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported!", http.StatusInternalServerError)
		return
	}

	messageChannel := make(chan TimerEvent)
	hub.AddClient <- messageChannel
	notify := w.(http.CloseNotifier).CloseNotify()

	setStreamHeaders(w)

	for i := 0; i < 1440; {
		select {
		case timerEvent := <-messageChannel:
			pushMessage(timerEvent, w, f)
		case <-time.After(time.Second * 3):
			t := TimerEvent{
				Time:  "-:--",
				Color: "#9E9E9E",
			}
			pushMessage(t, w, f)
			i++
		case <-notify:
			f.Flush()
			i = 1440
			hub.RemoveClient <- messageChannel
		}
	}
}

func pushMessage(t TimerEvent, w http.ResponseWriter, f http.Flusher) {
	jsonData, _ := json.Marshal(t)
	str := string(jsonData)
	fmt.Fprintf(w, eventWrapper, str)
	f.Flush()
}

func Handler() http.Handler {

	hub.Init()

	go countdown()

	r := mux.NewRouter()
	r.HandleFunc("/timer-events", timerEventsHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir("./assets/")))

	return r

}

func countdown() {
	for len(hub.Clients) == 0 {
		time.Sleep(time.Second)
	}
	r := time.Second
	t := timer.NewTimer(30*time.Second, r)
	t.Start()
	for d := range t.TickChannel() {
		hub.Messages <- TimerEvent{
			Time:  durationString(d, r),
			Color: thresholdColor(d),
		}
	}
}

func thresholdColor(d time.Duration) string {
	if d == 0 {
		return "#F44336"
	} else if d <= 30*time.Second {
		return "#FFEB3B"
	}
	return "#009688"
}

func durationString(d time.Duration, r time.Duration) string {
	const fmtSeconds = "%01d:%02d"
	const fmtMilliseconds = fmtSeconds + ".%03d"
	d = d.Round(r)
	m := d / time.Minute
	d -= m * time.Minute
	s := d / time.Second
	switch r {
	case time.Second:
		return fmt.Sprintf(fmtSeconds, m, s)
	case time.Millisecond:
		d -= s * time.Second
		ms := d / time.Millisecond
		return fmt.Sprintf(fmtMilliseconds, m, s, ms)
	default:
		panic("unsuported duration used for string conversion")
	}
}
