package main

import (
	"log"
	"os"

	"github.com/joho/godotenv"
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
