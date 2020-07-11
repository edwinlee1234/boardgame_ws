package main

import (
	"sync"

	log "github.com/sirupsen/logrus"
)

// Hub 放連線的地方
type Hub struct {
	ID               int32
	clients          map[*Client]bool
	register         chan *Client
	unregister       chan *Client
	broadcast        chan []byte
	destroy          chan bool
	FindClientByUUID chan string
	GetClient        chan *Client
	m                *sync.Mutex
}

func newHub(ID int32) *Hub {
	return &Hub{
		ID:               ID,
		clients:          make(map[*Client]bool),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		broadcast:        make(chan []byte),
		destroy:          make(chan bool),
		FindClientByUUID: make(chan string),
		GetClient:        make(chan *Client),
		m:                &sync.Mutex{},
	}
}

func (h *Hub) Run() {
	for {
		select {
		// 註冊新ws連線
		case client := <-h.register:
			log.WithFields(log.Fields{
				"uuid": client.UUID,
			}).Info("Hub register client")

			h.clients[client] = true

		// 取消ws連線
		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				log.WithFields(log.Fields{
					"uuid": client.UUID,
				}).Info("Hub unregister client")

				close(client.send)
				h.m.Lock()
				delete(h.clients, client)
				h.m.Unlock()
			}

		// 推播
		case msg := <-h.broadcast:
			for client, ok := range h.clients {
				if ok {
					client.send <- msg
				}
			}

		// 找client連線
		case UUID := <-h.FindClientByUUID:
			for client, isOk := range h.clients {
				if client.UUID == UUID && isOk {
					h.GetClient <- client
				}
			}

		// hub被刪掉，client連線全部刪
		case <-h.destroy:
			for client, ok := range h.clients {
				if ok {
					close(client.send)
					h.m.Lock()
					delete(h.clients, client)
					h.m.Unlock()
				}
			}
			return // END
		}
	}
}
