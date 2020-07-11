package main

import (
	"sync"
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// 80秒 ping一次hub，有沒有人在連，沒有就kill掉
	pingHubPeriod = 80 * time.Second
	// lobbyhub id
	lobby = 1
	// global caht
	globalChat = 2
)

// 永遠都要存在的hub
var vipHub = map[int32]bool{
	lobby:      true,
	globalChat: true,
}

func newWSCenter() *WSCenter {
	return &WSCenter{
		make(map[int32]*Hub),
		&sync.Mutex{},
	}
}

// WSCenter 裝hub的地方
type WSCenter struct {
	hubs map[int32]*Hub
	m    *sync.Mutex
}

// Init 把初始化的hub都new出來
func (g *WSCenter) Init() {
	for channelID, isOpen := range vipHub {
		if isOpen {
			hub := newHub(channelID)
			g.RegisterHub(channelID, hub)

			go hub.Run()
		}
	}
}

// FindHub 用ID去找hub
func (g *WSCenter) FindHub(ID int32) *Hub {
	if _, ok := g.hubs[ID]; !ok {
		return nil
	}

	return g.hubs[ID]
}

// RegisterHub RegisterHub
func (g *WSCenter) RegisterHub(ID int32, hub *Hub) error {
	log.WithFields(log.Fields{
		"id": ID,
	}).Info("RegisterHub")

	g.hubs[ID] = hub

	return nil
}

// UnregisterHub UnregisterHub
func (g *WSCenter) UnregisterHub(ID int32) {
	if _, ok := g.hubs[ID]; !ok {
		log.WithFields(log.Fields{
			"id": ID,
		}).Warning("刪hub ERROR，找不到hub id")

		return
	}

	hub := g.hubs[ID]
	hub.destroy <- true

	g.m.Lock()
	delete(g.hubs, ID)
	g.m.Unlock()
}

// HubCleaner 每一段時間去檢查全部的hub，是不是沒人在連線了，把它kill掉把記憶體釋出
func (g *WSCenter) HubCleaner() {
	ticker := time.NewTicker(60 * time.Second)

	for {
		select {
		case <-ticker.C:
			for ID, hub := range g.hubs {
				// 不是vip的hub，才需要檢查
				if _, isVip := vipHub[ID]; !isVip {
					if len(hub.clients) == 0 {
						g.UnregisterHub(ID)
					}
				}
			}
		}
	}
}
