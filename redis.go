package main

import (
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	uuid "github.com/satori/go.uuid"
	redistore "gopkg.in/boj/redistore.v1"
)

func connectRedisStore() {
	err := godotenv.Load()
	if err != nil {
		log.Panic(err)
	}

	host := os.Getenv("REDIS_HOST")
	password := os.Getenv("REDIS_PASSWORD")
	port := os.Getenv("REDIS_PORT")

	store, err = redistore.NewRediStore(10, "tcp", host+":"+port, password, key)
	if err != nil {
		panic(err)
	}
}

func getUUID(w http.ResponseWriter, r *http.Request) string {
	session, _ := store.Get(r, "userInfo")
	userUUID, ok := session.Values["uuid"].(string)

	// 如果session沒有，就new一個新的
	if !ok {
		UUID := uuid.Must(uuid.NewV4(), nil)

		userUUID = UUID.String()
		session.Values["uuid"] = userUUID
		session.Save(r, w)
	}

	return userUUID
}
