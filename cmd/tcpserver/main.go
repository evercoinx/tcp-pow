package main

import (
	"fmt"
	"os"

	"github.com/evercoinx/tcp-pow-server/internal/tcpserver"
)

func main() {
	if err := tcpserver.Start(":8000"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
