package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"syscall"
	"time"
)

type Connections struct {
	clients      map[chan string]bool
	addClient    chan chan string
	removeClient chan chan string
	messages     chan string
}

var hub = &Connections{
	clients:      make(map[chan string]bool),
	addClient:    make(chan (chan string)),
	removeClient: make(chan (chan string)),
	messages:     make(chan string),
}

func (hub *Connections) Init() {
	go func() {
		for {
			select {
			case s := <-hub.addClient:
				hub.clients[s] = true
				log.Println("Added new client")
			case s := <-hub.removeClient:
				delete(hub.clients, s)
				log.Println("Removed client")
			case msg := <-hub.messages:
				for s, _ := range hub.clients {
					s <- msg
				}
				log.Printf("Broadcast \"%v\" to %d clients", msg, len(hub.clients))
			}
		}
	}()
}

func httpHandler(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported!", http.StatusInternalServerError)
		return
	}

	if r.URL.Path == "/send" {
		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			log.Println(err)
		}
		str := string(body)
		hub.messages <- str
		f.Flush()
		return
	}

	messageChannel := make(chan string)
	hub.addClient <- messageChannel
	notify := w.(http.CloseNotifier).CloseNotify()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	for i := 0; i < 1440; {
		select {
		case msg := <-messageChannel:
			jsonData, _ := json.Marshal(msg)
			str := string(jsonData)
			if r.URL.Path == "/events/sse" {
				fmt.Fprintf(w, "data: {\"str\": %s, \"time\": \"%v\"}\n\n", str, time.Now())
			}
			f.Flush()
		case <-time.After(time.Second * 60):
			if r.URL.Path == "/events/sse" {
				fmt.Fprintf(w, "data: {\"str\": \"No Data\"}\n\n")
			}
			f.Flush()
			i++
		case <-notify:
			f.Flush()
			i = 1440
			hub.removeClient <- messageChannel
		}
	}
}

func countdown(d time.Duration) {
	for len(hub.clients) == 0 {
		time.Sleep(1 * time.Second)
	}
	for countup := 5; countup >= 0; countup-- {
		hub.messages <- strconv.Itoa(countup)
		time.Sleep(1 * time.Second)
	}
	for d >= 0 {
		hub.messages <- d.String()
		d -= time.Millisecond
		time.Sleep(time.Millisecond)
	}
}

func main() {
	fmt.Printf("application started at: %s\n", time.Now().Format(time.RFC822))
	var starttime int64 = time.Now().Unix()
	runtime.GOMAXPROCS(8)

	hub.Init()

	http.HandleFunc("/send", httpHandler)
	http.HandleFunc("/events/sse", httpHandler)
	http.HandleFunc("/events/lp", httpHandler)
	http.Handle("/", http.FileServer(http.Dir("./")))

	go http.ListenAndServe(":8000", nil)

	go countdown(time.Second * 30)

	var input string
	for input != "exit" {
		_, _ = fmt.Scanf("%v", &input)
		if input != "exit" {
			switch input {
			case "", "0", "5", "help", "info":
				fmt.Print("you can type \n1: \"exit\" to kill this application")
				fmt.Print("\n2: \"clients\" to show the amount of connected clients")
				fmt.Print("\n3: \"system\" to show info about the server")
				fmt.Print("\n4: \"time\" to show since when this application is running")
				fmt.Print("\n5: \"help\" to show this information")
				fmt.Println()
			case "1", "exit", "kill":
				fmt.Println("application get killed in 5 seconds")
				input = "exit"
				time.Sleep(5 * time.Second)
			case "2", "clients":
				fmt.Printf("connected to %d clients\n", len(hub.clients))
			case "3", "system":
				fmt.Printf("CPU cores: %d\nGo calls: %d\nGo routines: %d\nGo version: %v\nProcess ID: %v\n", runtime.NumCPU(), runtime.NumCgoCall(), runtime.NumGoroutine(), runtime.Version(), syscall.Getpid())
			case "4", "time":
				fmt.Printf("application running since %d minutes\n", (time.Now().Unix()-starttime)/60)
			}
		}
	}
	os.Exit(0)
}
