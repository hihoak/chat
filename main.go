package main

import (
	"github.com/hihoak/chat-app/chat"
	"log"
)

func main() {
	server, err := chat.NewServer()
	if err != nil {
		log.Fatal("failed to start server: ", err)
	}

	if runErr := server.Run(); runErr != nil {
		log.Fatal("failed to run server: ", runErr)
	}
}
