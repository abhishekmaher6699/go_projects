package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/gorilla/websocket"
)

type ClientMessage struct {
	Type      string `json:"type"`
	Recipient string `json:"recipient"`
	Content   string `json:"content"`
}

type ServerMessage struct {
	ID        uint64 `json:"id"`
	Type      string `json:"type"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Content   string `json:"content"`
}

func main() {

	user := flag.String("user", "alice", "username")
	recipient := flag.String("to", "bob", "recipient username")

	flag.Parse()

	u := url.URL{
		Scheme:   "ws",
		Host:     "localhost:8080",
		Path:     "/ws",
		RawQuery: "user=" + *user,
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	log.Println("connected as", *user)

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Println("connection closed:", err)
				return
			}

			var msg ServerMessage

			if err := json.Unmarshal(message, &msg); err != nil {
				log.Println("invalid message:", err)
				continue
			}

			switch msg.Type {
				case "message":
					log.Printf("[%d] %s: %s", msg.ID, msg.Sender, msg.Content)
				case "error":
					log.Printf("ERROR: %s", msg.Content)
			}
		}
	}()

	reader := bufio.NewReader(os.Stdin)

	for {
		text, err := reader.ReadString('\n')
		if err != nil {
			log.Println("failed to read input:", err)
			break
		}
		text = strings.TrimSpace(text)

		msg := ClientMessage{
			Type:      "message",
			Recipient: *recipient,
			Content:   text,
		}

		data, err := json.Marshal(msg)
		if err != nil {
			log.Println("failed to encode message:", err)
			continue
		}

		err = conn.WriteMessage(websocket.TextMessage, data)
		if err != nil {
			log.Println("failed to send message:", err)
			break
		}

	}
}
