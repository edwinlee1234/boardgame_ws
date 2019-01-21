package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"

	ErrorManner "boardgame_ws/error"

	uuid "github.com/satori/go.uuid"
)

// WsCheckRes 檢查channel的api回傳格式
type WsCheckRes struct {
	Status string                 `json:"status"`
	Error  map[string]interface{} `json:"error"`
}

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

func wsInstance(w http.ResponseWriter, r *http.Request) {
	channelParams := r.URL.Query()["channel"]
	if len(channelParams) < 1 {
		ErrorManner.ErrorRespone(errors.New("chanel error"), CHANNEL_ERROR, w, 400)
		return
	}

	// 判斷頻道開關
	channel := channelParams[0]
	req, err := http.Get(ServerAPI + "/api/checkChannel?channel=" + channel)
	if err != nil {
		ErrorManner.ErrorRespone(errors.New("Unexpect error"), UNEXPECT_ERROR, w, 500)
		return
	}

	contents, _ := ioutil.ReadAll(req.Body)
	var s = WsCheckRes{}
	err = json.Unmarshal(contents, &s)
	if err != nil {
		ErrorManner.ErrorRespone(errors.New("Unexpect error"), UNEXPECT_ERROR, w, 500)
		return
	}

	if s.Status != "success" {
		ErrorManner.ErrorRespone(errors.New("chanel error"), CHANNEL_ERROR, w, 400)
		return
	}

	var channelID int32
	channelIDArrs := r.URL.Query()["id"]
	if len(channelIDArrs) >= 1 {
		var err error
		channelIDArr := channelIDArrs[0]
		channelIDInt, err := strconv.Atoi(channelIDArr)
		channelID = int32(channelIDInt)

		if err != nil {
			log.Println("create channel error")
			return
		}
	}

	if channel == "lobby" {
		channelID = 1
	}

	userUUID := getUserUUID(w, r)

	// 連線ws
	ConnWs(channelID, userUUID, w, r)
}

// CheckAllChannel 把現在的hub & client都println出來
func CheckAllChannel() {
	for hub, boolan := range group.hubs {
		log.Println("hub:")
		log.Println(hub.id)
		log.Println(boolan)
		for address, boolan := range hub.clients {
			log.Println(address)
			log.Println(boolan)
		}
	}
}

// CreateLobby 開server的時候就會create一個lobby的hub
func CreateLobby() {
	hub := NewHub(LobbyID)
	go hub.Run()

	group.addHubChan <- hub
}

// ConnWs 連線websocket
func ConnWs(channelID int32, UUID string, w http.ResponseWriter, r *http.Request) {
	var hub *Hub

	// 去Group搜尋hub
	group.findHubChan <- channelID
	hub = <-groupFindHubChan

	log.Println("hub", hub)
	// 如果Group沒有這個hub，新增一個
	if hub == nil {
		flag.Parse()
		hub = NewHub(channelID)
		go hub.Run()

		group.addHubChan <- hub
	}

	// 新增Client
	serveWs(UUID, hub, w, r)
}

// BroadcastChannel 推播某頻道
func BroadcastChannel(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var data BroadcastRequest
	err := decoder.Decode(&data)

	if err != nil {
		log.Println(err)
		return
	}

	group.findHubChan <- data.ChannelID
	hub := <-groupFindHubChan

	if hub == nil {
		log.Println("不存在這hub id: ", data.ChannelID)
		return
	}
	// 格或處理
	message := bytes.TrimSpace(bytes.Replace(data.Data, newline, space, -1))
	// 針對hub裡面的連線都推播
	hub.broadcast <- message
}

// BroadcastUser 推播單一個user
func BroadcastUser(w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var data BroadcastUserRequest
	err := decoder.Decode(&data)

	if err != nil {
		log.Println(err)
		return
	}

	// 找hub
	group.findHubChan <- data.ChannelID
	hub := <-groupFindHubChan
	if hub == nil {
		log.Println("hun nil id:", data.ChannelID)
		return
	}

	// 找user
	hub.findClientChan <- data.UUID
	oldClient := <-hubFindClientChan

	if oldClient == nil {
		log.Println("user nil id:", data.UUID)
		return
	}

	// 推播
	message := bytes.TrimSpace(bytes.Replace(data.Data, newline, space, -1))
	oldClient.send <- message
}

// 取得UUID
func getUserUUID(w http.ResponseWriter, r *http.Request) string {
	session, _ := store.Get(r, "userInfo")
	// 用string的格式取出來
	// *這個用法很重要
	userUUID, ok := session.Values["uuid"].(string)

	// 如果session沒有，就new一個新的
	if !ok {
		UUID := uuid.Must(uuid.NewV4())
		// UUID轉成string
		userUUID = UUID.String()
		session.Values["uuid"] = userUUID
		session.Save(r, w)
	}

	return userUUID
}
