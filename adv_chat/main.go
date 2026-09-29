package main

import (
	"log"
	"net/http"
	"advchat/server"
)

func main() {

	db, err := server.NewDatabase("chat.db")
	if err != nil {
		log.Fatal(err)
	}

	if err := db.Init(); err != nil {
		log.Fatal(err)
	}

	chatServer  := server.NewChatServer(db)

	http.HandleFunc("/ws", chatServer.HandleWebSocket)

	log.Println("chat server running on :8080")

	err = http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal(err)
	}
}