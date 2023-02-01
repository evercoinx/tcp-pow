package main

import (
	"os"
	"time"

	"github.com/evercoinx/tcp-pow-server/internal/tcpclient"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
)

type AppConfig struct {
	TCPServer TCPServer `yaml:"tcp_server"`
}

type TCPServer struct {
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

	if err := tcpclient.QueryPipeline(config.TCPServer.Address); err != nil {
		log.Fatal(err)
	}
}
