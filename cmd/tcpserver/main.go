package main

import (
	"os"
	"time"

	"github.com/evercoinx/tcp-pow/internal/tcpserver"
	"github.com/jinzhu/configor"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

type AppConfig struct {
	TCPServer TCPServer `yaml:"tcp_server"`
	Redis     Redis     `yaml:"redis"`
}

type TCPServer struct {
	Address         string        `yaml:"address" default:"0.0.0.0:8000"`
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
	var cfg AppConfig
	configor.Load(&cfg, "./config/config.yml")

	redisCli := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.Address,
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})

	srv := tcpserver.NewServer(redisCli, cfg.TCPServer.CacheExpiration)
	if err := srv.Start(cfg.TCPServer.Address); err != nil {
		log.Fatal(err)
	}
}
