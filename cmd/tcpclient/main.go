package main

import (
	"os"
	"time"

	"github.com/evercoinx/tcp-pow/internal/tcpclient"
	"github.com/jinzhu/configor"
	log "github.com/sirupsen/logrus"
)

type AppConfig struct {
	TCPClient TCPClient `yaml:"tcp_client"`
}

type TCPClient struct {
	Address      string        `yaml:"address" default:"127.0.0.1:8000"`
	WaitInterval time.Duration `yaml:"wait_interval" default:"10s"`
}

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true, TimestampFormat: time.StampMilli})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	var cfg AppConfig
	configor.Load(&cfg, "./config/config.yml")

	for {
		if err := tcpclient.QueryPipeline(cfg.TCPClient.Address); err != nil {
			log.Fatal(err)
		}
		time.Sleep(cfg.TCPClient.WaitInterval)
	}
}
