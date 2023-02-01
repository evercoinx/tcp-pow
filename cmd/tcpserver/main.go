package main

import (
	"os"

	"github.com/evercoinx/tcp-pow-server/internal/tcpserver"
	log "github.com/sirupsen/logrus"
)

func init() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetOutput(os.Stdout)
	log.SetLevel(log.DebugLevel)
}

func main() {
	if err := tcpserver.Start(":8000"); err != nil {
		log.Fatal(err)
	}
}
