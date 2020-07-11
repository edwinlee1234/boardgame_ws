package main

import (
	"bytes"
	"encoding/json"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// BroadcastRequest 推播的request格式
type BroadcastRequest struct {
	ChannelID int32  `json:"channed_id"`
	Data      []byte `json:"data"`
}

// BroadcastUserRequest 推播單一user的request格式
type BroadcastUserRequest struct {
	ChannelID int32  `json:"channed_id"`
	UUID      string `json:"UUID"`
	Data      []byte `json:"data"`
}

// BroadcastChannel 推播某頻道
// TODO 加一個安全機制，不是每一個人都可以自已推這個頻道
func BroadcastChannel(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var data BroadcastRequest
	err := decoder.Decode(&data)

	if err != nil {
		log.Println(err)
		return
	}

	hub := wsCenter.FindHub(data.ChannelID)
	if hub == nil {
		log.WithFields(log.Fields{
			"ChannelID": data.ChannelID,
			"API":       "BroadcastChannel",
		}).Info("Hub Empty")

		return
	}

	// 格或處理
	message := bytes.TrimSpace(bytes.Replace(data.Data, newline, space, -1))
	// 針對hub裡面的連線都推播
	hub.broadcast <- message
}

// BroadcastUser 推播單一個user
// TODO 加一個安全機制，不是每一個人都可以自已推這個頻道
func BroadcastUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var data BroadcastUserRequest
	err := decoder.Decode(&data)

	if err != nil {
		log.Println(err)
		return
	}

	// 找hub
	hub := wsCenter.FindHub(data.ChannelID)
	if hub == nil {
		log.WithFields(log.Fields{
			"ChannelID": data.ChannelID,
			"API":       "BroadcastUser",
		}).Info("Hub Empty")

		return
	}

	// 找user
	hub.FindClientByUUID <- data.UUID
	oldClient := <-hub.GetClient
	if oldClient == nil {
		log.WithFields(log.Fields{
			"UUID":      data.UUID,
			"ChannelID": data.ChannelID,
			"API":       "BroadcastUser",
		}).Info("Hub user nil")

		return
	}

	// 推播
	message := bytes.TrimSpace(bytes.Replace(data.Data, newline, space, -1))
	oldClient.send <- message
}
