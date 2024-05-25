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

	"github.com/google/uuid"
)

// JSON-RPC Client for sending and receiving messages - JSON-RPC is transport agnostic.
type JsonRpcClient struct {
	transport io.ReadWriter
}

// Build a JSON-RPC request
func (c *JsonRpcClient) BuildRequest(params any, method string) JsonRpcRequest {
	requestId := uuid.New().String()
	request := JsonRpcRequest{Id: requestId, JsonRpc: JsonRpcVersion, Method: method, Params: params}
	return request
}

// Send a JSON-RPC Request
func (c *JsonRpcClient) Send(request JsonRpcRequest) error {
	requestJson, err := json.Marshal(request)
	if err != nil {
		log.Println(err)
		return err
	}

	log.Printf("Sending json-rpc request: [%s] \n", request.Method)
	_, err = c.transport.Write(requestJson)

	if err != nil {
		log.Printf("Failed to send json rpc request: [id=%s, method=%s] \n", request.Id, request.Method)
		log.Println(err)
		return err
	}

	log.Println("Request succesfully sent")
	return nil
}

// TODO: Implement timeout, context? Pass Buffer size into recv?
// Receive a JSON-RPC Response containing a result or error
func (c *JsonRpcClient) Recv() (JsonRpcResponse, error) {
	buf := make([]byte, 1024)
	_, err := c.transport.Read(buf)

	if err != nil {
		log.Println("Failed reading json-rpc response", err)
		return JsonRpcResponse{}, err
	}

	response := JsonRpcResponse{}
	err = json.Unmarshal(buf, &response)

	if err != nil {
		log.Println("Failed to deserialize response", err)
		return JsonRpcResponse{}, err
	}
	return response, nil
}

// Send a JSON-RPC Request and read the JSON-RPC Response from the server
func (c *JsonRpcClient) SendAndRecv(request JsonRpcRequest) JsonRpcResponse {
	err := c.Send(request)
	if err != nil {
		panic("Failed to send")
	}
	response, err := c.Recv()
	if err != nil {
		panic("Failed to recv....")
	}
	return response
}

// Handle JSON-RPC Notifications (server push)
func (c *JsonRpcClient) HandleNotifications() {
	buf := make([]byte, 1024)
	for {
		_, err := c.transport.Read(buf)
		if err != nil {
			log.Println("Error reading reponse from server:", err)
		}
		log.Printf("Response from server: [%s] \n", buf)
		clear(buf)
	}
}

// Send a request to chat
func (c *JsonRpcClient) SendChatRequest(msg []byte) JsonRpcResponse {
	params := ChatRequestParams{Msg: msg}
	request := c.BuildRequest(params, ChatRpcMethod)
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

func main() {
	tcpConnection := TCPConnect("localhost", 8080)
	defer tcpConnection.Close()
	client := &JsonRpcClient{transport: tcpConnection}

	// Seperate go-routine for handling JSON-RPC Notifications
	go client.HandleNotifications()

	// TODO: Factor this into a func.
	scanner := bufio.NewScanner(os.Stdin)

	for {
		scanner.Scan()
		msg := scanner.Text()

		if strings.ToLower(msg) == "exit" {
			log.Println("Exiting chat room")
			break
		}
		client.SendChatRequest([]byte(msg))
	}
}
