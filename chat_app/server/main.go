package main

import (
	"bufio"
	"fmt"
	"net"
	"sync"
)

func handleClient(conn net.Conn, clients map[string]net.Conn, mu *sync.Mutex) {
	defer conn.Close()

	clientID := conn.RemoteAddr().String()

	mu.Lock()
	clients[clientID] = conn
	mu.Unlock()

	fmt.Println("Client connected:", clientID)
	fmt.Println("Connected clients:", len(clients))

	scanner := bufio.NewScanner(conn)

	if !scanner.Scan() {
		mu.Lock()
		delete(clients, clientID)
		mu.Unlock()
		return
	}

	username := scanner.Text()
	fmt.Println("User joined:", username)



	for scanner.Scan() {
		message := scanner.Text()

		fmt.Printf("%s: %s\n", username, message)

		formattedMsg := fmt.Sprintf(
			"%s: %s\n",
			username,
			message,
		) 

		broadcast([]byte(formattedMsg), clientID, clients, mu)
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading;", err)
	}

	mu.Lock()
	delete(clients, clientID)
	mu.Unlock()

	fmt.Println("Client disconnected:", clientID)



}

func broadcast(
	message []byte,
	sender string,
	clients map[string]net.Conn,
	mu *sync.Mutex,
) {
	mu.Lock()
	defer mu.Unlock()

	for clientID, conn := range clients {
		if clientID == sender {
			continue
		}

		conn.Write(message)
	}
}

func main() {
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		panic(err)
	}
	defer listener.Close()

	var mu sync.Mutex
	clients := make(map[string]net.Conn)

	fmt.Println("Chat server listening on port 8080...")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting client:", err)
			continue
		}

		go handleClient(conn, clients, &mu)
	}
}
