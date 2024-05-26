package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"

	"github.com/google/uuid"
)

/*   Chat-room server (IRC analog) - First iteration
 *
 *   Connection requirements:
 *   - Server can handle multiple connections concurrently
 *   - Server can fan-out messages to all connected clients
 *   - Server will handle when the client disconnects (discard active client session)
 *   - Start serial impl first, then parallelize
 *
 *  Chat-room client requirements:
 *  - Clients can connect to the chat room server via a TCP socket
 *  - Can send messages via stdinput
 *  - Can receive messages
 *  - Need some unique name on connect (UserId) or something...
 *
 *  Server JSON-RPC Methods.... [sendChatMsg, createUser, createChatRoom, joinChatRoom, leaveChatRoom, listChatRooms, listUsers?, PrivateMessageUser]
 *                                req     req         req            req      req            req
 *
 *  Transport messages following JSON-RPC over TCP sockets...
 *    Pub/sub for notifications... Subscribe to chat ROOM when joined.
 *
 *  Server maintains a message history using a stack ~_~
 *  TODO: Add wrapped errors for JSON-RPC error codes and message
 */

// Should the dispatcher have the connection manager struct?

// TODO: Have some sort of config struct + yaml file to read in configuration as well

// TCP Server
type Server struct {
	port              int
	listener          net.Listener
	connectionService *ConnectionService
	dispatcher        *JsonRpcDispatcher
}

// Start the server and listen for connections (Main entry point)
func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))

	if err != nil {
		log.Fatalf("Failed to start server! Exiting", err)
	}

	s.listener = listener
	log.Printf("Server starting... Listening on port: [%d] \n", s.port)
	s.Listen()
}

// Listen for new connections - each connection will spawn a goroutine
func (s *Server) Listen() {
	for {
		conn, err := s.listener.Accept()
		log.Println("New connection accepted")
		if err != nil {
			log.Println("Error accepting incoming connection:", err)
			continue
		}
		connectionId := uuid.New().String()
		log.Printf("Connection Accepted. ConnectionId = [%s]\n", connectionId)
		connection := &Connection{id: connectionId, Conn: conn}
		go s.HandleConnectionMessages(connection)
	}
}

// Hanle incoming messages from a connection
func (s *Server) HandleConnectionMessages(connection *Connection) {
	buf := make([]byte, 1024)
	var request JsonRpcRequest

	for {
		bytesRead, err := connection.Read(buf)
		if err != nil {
			log.Fatalln("Error reading connection buffer", err.Error())
		}

		log.Printf("Read [%d] bytes from connection [%s} \n", bytesRead, connection.id)
		log.Printf("Message: %s\n", buf)

		err = json.Unmarshal(buf, &request)
		if err != nil {
			log.Printf("Failed deserializing message [%s] into JSON-RPC Request\n", buf)
			connection.Write([]byte("Invalid Request, must follow JSON-RPC Request schema"))
			continue
		}

		s.dispatcher.Dispatch(request, connection)
		clear(buf)
	}
}

func ChatMessageHandler(request JsonRpcRequest) JsonRpcResponse {
	var params ChatRequestParams
	err := json.Unmarshal(request.Params, &params)
	if err != nil {
		errorResponse := JsonRpcResponse{Id: request.Id, JsonRpc: request.JsonRpc, Error: &JsonRpcError{Code: -32600, Message: "Invalid Request"}}
		return errorResponse
	}
	response := JsonRpcResponse{Id: request.Id, JsonRpc: request.JsonRpc, Result: SuccessResult{Success: true}}
	return response
}

func main() {
	dispatcher := NewDispatcher()
	dispatcher.AddMethod("chat", ChatMessageHandler)
	connectionService := &ConnectionService{store: NewConnectionStore()}
	server := &Server{port: 8080, connectionService: connectionService, dispatcher: dispatcher}
	server.Start()
}
