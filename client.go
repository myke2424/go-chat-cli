package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

func Connect(host string, port int) net.Conn {
	address := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Connecting to server [%s]\n", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalln("Failed to connect to server!", err)
	}
	log.Println("Connected")
	return conn
}

func SendMessage(msg []byte, conn net.Conn) {
	log.Printf("Sending message to server: [%s]\n", string(msg))
	_, err := conn.Write(msg)
	if err != nil {
		log.Fatalln("Failed to send message to server", err)
	}
	log.Println("Message sent")
}

func HandleServerMessages(conn net.Conn) {
	for {
		response, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			log.Println("Error reading response from server:", err)
		}
		log.Println("Response from server: ", response)
	}
}

func main() {
	conn := Connect("localhost", 8080)
	defer conn.Close()
	go HandleServerMessages(conn) // Seperate goroutine to handle messages from the server

	// Reading from stdin is the "UI" layer in this case, refactor this to make it more flexible
	// Need two goroutines, one for sending and one for receiving messages.
	scanner := bufio.NewScanner(os.Stdin)

	for {
		scanner.Scan()
		msg := scanner.Text()

		if strings.ToLower(msg) == "exit" {
			log.Println("Exiting chat room")
			break
		}
		SendMessage([]byte(msg), conn)

	}
}
