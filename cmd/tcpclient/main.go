package main

import (
	"os"
	"time"

	"github.com/evercoinx/go-tcp-pow/internal/tcpclient"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
)

type AppConfig struct {
	TCPClient TCPClient `yaml:"tcp_client"`
}

type TCPClient struct {
	Address string `yaml:"address" default:"127.0.0.1:8000"`
}

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, TimestampFormat: time.StampMilli})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	var config AppConfig
	configor.Load(&config, "./config/config.yml")

	if err := tcpclient.QueryPipeline(config.TCPClient.Address); err != nil {
		log.Fatal(err)
	}
}
