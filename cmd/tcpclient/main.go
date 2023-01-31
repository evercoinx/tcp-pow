package main

import (
	"fmt"
	"os"

	"github.com/evercoinx/tcp-pow-server/internal/tcpclient"
)

func main() {
	if err := tcpclient.Query(":8000"); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
