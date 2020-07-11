package main

import (
	"errors"
	"net/http"
	"strconv"

	log "github.com/sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	redistore "gopkg.in/boj/redistore.v1"
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	//如果有 cross domain 的需求，可加入這個，不檢查 cross domain
	CheckOrigin: func(r *http.Request) bool { return true },
}

var (
	key        = []byte("asEg$1#jssSf245")
	store      *redistore.RediStore
	wsCenter   *WSCenter
	statistics *Statistics
)

func init() {
	// 統計物件
	statistics = newStatistics()
	go statistics.Run()

	wsCenter = newWSCenter()
	wsCenter.Init()
	go wsCenter.HubCleaner()

	connectRedisStore()
}

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

func main() {
	r := mux.NewRouter()
	r.Use(before)

	r.HandleFunc("/ws", ws).Methods("GET", "OPTIONS")
	r.HandleFunc("/test", test).Methods("GET")
	r.HandleFunc("/broadcast", BroadcastChannel).Methods("POST", "OPTIONS")
	r.HandleFunc("/broadcastUser", BroadcastUser).Methods("POST", "OPTIONS")

	log.Println("server start at :8000")
	if err := http.ListenAndServe(":8000", r); err != nil {
		log.Fatal(err)
	}
}

func test(w http.ResponseWriter, r *http.Request) {
}

// 連接ws
func ws(w http.ResponseWriter, r *http.Request) {
	p := r.URL.Query()
	channel := p.Get("channel")
	idstr := p.Get("id")

	if channel == "" || idstr == "" {
		ErrorRespone(errors.New("miss params"), PARAMS_ERROR, w, 400)
		return
	}

	if enable, exist := channelSupport[channel]; !exist || !enable {
		ErrorRespone(errors.New("chanel error"), CHANNEL_ERROR, w, 400)
		return
	}

	id64, err := strconv.ParseInt(idstr, 10, 32)
	if err != nil {
		ErrorRespone(err, PARAMS_ERROR, w, 400)
		return
	}

	id := int32(id64)

	hub := wsCenter.FindHub(id)
	if hub == nil {
		hub = newHub(id)
		go hub.Run()

		wsCenter.RegisterHub(id, hub)
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Error(err.Error())

		ErrorRespone(err, UNEXPECT_ERROR, w, 500)
		return
	}

	uuid := getUUID(w, r)

	client := &Client{
		uuid,
		conn,
		hub,
		make(chan []byte),
	}

	hub.register <- client

	go client.ReceiveClientMsgProcess()
	go client.SendClientMsgProcess()
}
