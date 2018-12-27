package main

import (
	"fmt"
	"log"
	"net/http"

	middleware "./middleware"
	"github.com/gorilla/mux"
	redistore "gopkg.in/boj/redistore.v1"
)

// ws的連線都會放在這邊
var group *Group

var (
	key   = []byte("super-secret-key")
	store *redistore.RediStore
)

func init() {
	CreateGroup()
	CreateLobby()
	connectRedisStore()
}

func main() {
	r := mux.NewRouter()
	r.Use(middleware.Before)

	// WS
	r.HandleFunc("/ws", wsInstance).Methods("GET", "OPTIONS")
	r.HandleFunc("/broadcast", BroadcastChannel).Methods("POST", "OPTIONS")
	r.HandleFunc("/broadcastUser", BroadcastUser).Methods("POST", "OPTIONS")

	// Test
	r.HandleFunc("/test", test).Methods("GET", "OPTIONS")

	err := http.ListenAndServe(":8000", r)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func test(w http.ResponseWriter, r *http.Request) {
	fmt.Println("test")
}
