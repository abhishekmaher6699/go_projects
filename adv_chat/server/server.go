package server

import (
	"encoding/json"
	"github.com/gorilla/websocket"
	"log"
	"net/http"
	"sync"
	"sync/atomic"
)


var messageId uint64

type Message struct {
	ID 	      uint64 `json:"id"`
	Type      string `json:"type"`
	Sender    string `json:"sender"`
	Recipient string `json:"recipient"`
	Content   string `json:"content"`
}

type Client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type ChatServer struct {
	clients map[string]*Client
	db *Database
	mu      sync.RWMutex
}

func (c *Client) Send(message Message) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return c.conn.WriteMessage(
		websocket.TextMessage,
		data,
	)
}

func (c *Client) SendError(content string) error {
	return c.Send(Message{
		Type:    "error",
		Content: content,
	})
}

func NewChatServer(db *Database) *ChatServer {
	return &ChatServer{
		clients: make(map[string]*Client),
		db: db,
	}
}

func nextMessageId() uint64 {
	return atomic.AddUint64(&messageId, 1)
}

func (s *ChatServer) SendMessage(msg Message) {

	s.mu.RLock()
	client, exists := s.clients[msg.Recipient]
	s.mu.RUnlock()

	if !exists {
		log.Println("recipient not connected:", msg.Recipient)
		return
	}

	if err := client.Send(msg); err != nil {
		log.Println("failed to deliver message:", err)
		return
	}

	log.Printf("%s -> %s: %s", msg.Sender, msg.Recipient, msg.Content)
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *ChatServer) HandleWebSocket(w http.ResponseWriter, r *http.Request) {

	userId := r.URL.Query().Get("user")

	if userId == "" {
		http.Error(w, "missing user", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	client := &Client{
		conn: conn,
	}

	s.mu.Lock()
	s.clients[userId] = client
	s.mu.Unlock()

	log.Println(userId, "connected")

	defer func() {
		s.mu.Lock()
		delete(s.clients, userId)
		s.mu.Unlock()

		conn.Close()

		log.Println(userId, "disconnected")
	}()

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			break
		}

		var msg Message

		if err := json.Unmarshal(message, &msg); err != nil {
			if err := client.SendError("invalid JSON"); err != nil {
				log.Println("failed to send error:", err)
			}
			continue
		}

		msg.Sender = userId
		msg.ID =  nextMessageId()

		if err := s.db.SaveMessage(msg); err != nil {
			log.Println("failed to save message:", err)
			continue
		}

		s.SendMessage(msg)
	}
}
