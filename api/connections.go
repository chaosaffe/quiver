package api

import "log"

type Connections struct {
	Clients      map[chan TimerEvent]bool
	AddClient    chan chan TimerEvent
	RemoveClient chan chan TimerEvent
	Messages     chan TimerEvent
}

func (hub *Connections) Init() {
	go func() {
		for {
			select {
			case s := <-hub.AddClient:
				hub.Clients[s] = true
				log.Println("Added new client")
			case s := <-hub.RemoveClient:
				delete(hub.Clients, s)
				log.Println("Removed client")
			case msg := <-hub.Messages:
				for s, _ := range hub.Clients {
					s <- msg
				}
				log.Printf("Broadcast \"%v\" to %d clients", msg, len(hub.Clients))
			}
		}
	}()
}
