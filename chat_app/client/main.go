package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	conn, err := net.Dial("tcp", "127.0.0.1:8080")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	fmt.Println("Connected to chat server!")

	go func() {
		buffer := make([]byte, 1024)

		for {
			n, err := conn.Read(buffer)
			if err != nil {
				fmt.Println("Disconnected")
				return
			}

			fmt.Print(string(buffer[:n]))
		}
	}()

	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Enter username: ")
	scanner.Scan()

	username := scanner.Text()

	_, err = fmt.Fprintf(conn, "JOIN|%s\n", username)
	if err != nil {
		fmt.Println("Error sending username:", err)
		return
	}

	for scanner.Scan() {
		message := scanner.Text()

		_, err := fmt.Fprintf(conn, "MSG|%s\n", message)
		if err != nil {
			fmt.Println("Error sending:", err)
			return
		}
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("Error reading input:", err)
	}

}
