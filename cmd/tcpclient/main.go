package main

import (
	"os"

	"github.com/evercoinx/tcp-pow-server/internal/tcpclient"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.InfoLevel)
}

func main() {
	if err := tcpclient.QueryPipeline("127.0.0.1:8000"); err != nil {
		log.Fatal(err)
	}
}
