package main

import (
	"os"
	"time"

	"github.com/evercoinx/tcp-pow-server/internal/tcpserver"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	redisCli := redis.NewClient(&redis.Options{
		Addr:     "127.0.0.1:6379",
		Password: "",
		DB:       0,
	})

	srv := tcpserver.NewServer(redisCli, 1*time.Minute)
	if err := srv.Start("127.0.0.1:8000"); err != nil {
		log.Fatal(err)
	}
}
