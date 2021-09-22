package main

import (
	"clappebruno/chatingo/client"
	"clappebruno/chatingo/server"
	"flag"
	"fmt"
	"strings"
)

const (
	IP   = "127.0.0.1" // IP local
	PORT = 3500        // Port utilisé
)

func main() {
	var mode string

	flag.StringVar(&mode, "mode", "client", "--mode client or --mode server")
	flag.Parse()

	switch strings.ToLower(mode) {
	case "server":
		server := server.New(IP, PORT)
		server.Launch()
	case "client":
		client := client.New(IP, PORT)
		client.Launch()
	default:
		fmt.Println("Usage : go run main.go --mode [server|client]")
	}
}
