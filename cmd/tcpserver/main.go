package main

import (
	"os"
	"time"

	"github.com/evercoinx/tcp-pow-server/internal/tcpserver"
	"github.com/jinzhu/configor"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

type AppConfig struct {
	TCPServer TCPServer `yaml:"tcp_server"`
	Redis     Redis     `yaml:"redis"`
}

type TCPServer struct {
	Address         string        `yaml:"address" default:"127.0.0.1:8000"`
	CacheExpiration time.Duration `yaml:"cache_expiration" default:"1m"`
}

type Redis struct {
	Address  string `yaml:"address" default:"127.0.0.1:6379"`
	Password string `yaml:"password" default:""`
	DB       int    `yaml:"db" default:"0"`
}

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, TimestampFormat: time.StampMilli})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	var config AppConfig
	configor.Load(&config, "./config/config.yml")

	redisCli := redis.NewClient(&redis.Options{
		Addr:     config.Redis.Address,
		Password: config.Redis.Password,
		DB:       config.Redis.DB,
	})

	srv := tcpserver.NewServer(redisCli, config.TCPServer.CacheExpiration)
	if err := srv.Start(config.TCPServer.Address); err != nil {
		log.Fatal(err)
	}
}
