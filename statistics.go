package main

import (
	"time"

	log "github.com/sirupsen/logrus"
)

const (
	// 30秒統計一次人數
	statisticsRunPeriod = 30 * time.Second
)

func newStatistics() *Statistics {
	return &Statistics{}
}

// Statistics ws統計物件
type Statistics struct{}

// Run Run
func (s *Statistics) Run() {
	logWsStatusTicker := time.NewTicker(statisticsRunPeriod)

	for {
		select {
		case <-logWsStatusTicker.C:
			s.logWsStatus()
		}
	}
}

// 把WSCenter裡面的hub人數跟ID都log出來
func (s *Statistics) logWsStatus() {
	if len(wsCenter.hubs) == 0 {
		log.Info("WSCenter ws empty")

		return
	}

	for ID, hub := range wsCenter.hubs {
		log.WithFields(log.Fields{
			"id":        ID,
			"connected": len(hub.clients),
		}).Info("WSCenter ws count: ")
	}
}
