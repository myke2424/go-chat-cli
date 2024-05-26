package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// JSON-RPC Client for sending and receiving messages - JSON-RPC is transport agnostic.
type JsonRpcClient struct {
	transport    io.ReadWriter
	responseChan map[string]chan JsonRpcResponse
	mu           sync.Mutex
}

// Build a JSON-RPC request
func (c *JsonRpcClient) BuildRequest(params []byte, method string) JsonRpcRequest {
	requestId := uuid.New().String()
	request := JsonRpcRequest{Id: requestId, JsonRpc: "2.0", Method: method, Params: params}

	c.mu.Lock()
	c.responseChan[requestId] = make(chan JsonRpcResponse, 1)
	c.mu.Unlock()
	return request
}

// Send a JSON-RPC Request
func (c *JsonRpcClient) Send(request JsonRpcRequest) error {
	requestJson, err := json.Marshal(request)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Sending json-rpc request: %s \n", string(requestJson))
	_, err = c.transport.Write(requestJson)

	if err != nil {
		log.Printf("Failed to send json rpc request: %s \n", string(requestJson))
		log.Println(err)
		return err
	}

	log.Println("Request successfully sent")
	return nil
}

// Send a JSON-RPC Request and read the JSON-RPC Response from the server
func (c *JsonRpcClient) SendAndRecv(request JsonRpcRequest) JsonRpcResponse {
	err := c.Send(request)
	if err != nil {
		panic("Failed to send json-rpc request")
	}

	c.mu.Lock()
	responseChan := c.responseChan[request.Id]
	c.mu.Unlock()

	timeout := time.After(5 * time.Second)
	select {
	case response := <-responseChan:
		if response.Id == request.Id {
			return response
		}
	case <-timeout:
		panic("Timeout waiting for response")
	}
	return JsonRpcResponse{}
}

// Handle Messages from the server
func (c *JsonRpcClient) HandleServerMessages() {
	decoder := json.NewDecoder(c.transport)
	for {
		var message json.RawMessage
		if err := decoder.Decode(&message); err != nil {
			if err == io.EOF {
				log.Println("Connection closed by server")
				return
			}
			log.Println("Error decoding message:", err)
			continue
		}

		// Determine if it's a notification or response
		var messageMap map[string]interface{}
		if err := json.Unmarshal(message, &messageMap); err != nil {
			log.Println("Error unmarshalling message:", err)
			continue
		}

		if _, ok := messageMap["method"]; ok {
			// Handle notification
			var notification JsonRpcNotification
			if err := json.Unmarshal(message, &notification); err != nil {
				log.Println("Error deserializing notification", err)
				continue
			}
			log.Printf("Notification from server: %s \n", formatJSON(notification))
		} else {
			// Handle response
			var response JsonRpcResponse
			if err := json.Unmarshal(message, &response); err != nil {
				log.Println("Error deserializing response", err)
				continue
			}
			log.Printf("Response from server: %s \n", formatJSON(response.Result))

			c.mu.Lock()
			if ch, ok := c.responseChan[response.Id]; ok {
				ch <- response
				close(ch)
				delete(c.responseChan, response.Id)
			}
			c.mu.Unlock()
		}
	}
}

// Send a request to chat
func (c *JsonRpcClient) SendChatRequest(msg []byte) JsonRpcResponse {
	params, _ := json.Marshal(ChatRequestParams{Msg: msg})
	request := c.BuildRequest(params, "chat")
	response := c.SendAndRecv(request)
	return response
}

// Connect to the chat server using a TCP socket
func TCPConnect(host string, port int) net.Conn {
	address := fmt.Sprintf("%s:%d", host, port)
	log.Printf("Connecting to server [%s]\n", address)
	conn, err := net.Dial("tcp", address)
	if err != nil {
		log.Fatalln("Failed to connect to TCP server!", err)
	}
	log.Println("Connected")
	return conn
}

func formatJSON(v interface{}) string {
	formatted, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("Error formatting JSON: %v", err)
	}
	return string(formatted)
}

func main() {
	tcpConnection := TCPConnect("localhost", 8080)
	defer tcpConnection.Close()
	client := &JsonRpcClient{transport: tcpConnection, responseChan: make(map[string]chan JsonRpcResponse)}

	go client.HandleServerMessages()
	scanner := bufio.NewScanner(os.Stdin)

	for {
		scanner.Scan()
		msg := scanner.Text()

		if strings.ToLower(msg) == "exit" {
			log.Println("Exiting chat room")
			break
		}
		go client.SendChatRequest([]byte(msg))
	}
}
