package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"

	"github.com/google/uuid"
)

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
		s.connectionService.AddConnection(connection)
		go s.HandleConnectionMessages(connection)
	}
}

// Handle incoming messages from a connection
func (s *Server) HandleConnectionMessages(connection *Connection) {
	buf := make([]byte, 1024)
	var request JsonRpcRequest

	for {
		connections := s.connectionService.ListConnections()
		log.Printf("[%d] connections listening", len(connections))

		bytesRead, err := connection.Read(buf)
		message := buf[:bytesRead]

		if err != nil {
			log.Fatalln("Error reading connection buffer", err.Error())
			if err == io.EOF {
				log.Printf("Connection closed: %s", connection.id)
				s.connectionService.DeleteConnection(connection.id)
				continue
			}

		}

		log.Printf("Read [%d] bytes from connection [%s} \n", bytesRead, connection.id)
		log.Printf("Message: %s\n", buf)

		err = json.Unmarshal(message, &request)
		if err != nil {
			log.Printf("Failed deserializing message [%s] into JSON-RPC Request\n", buf)
			connection.Write([]byte("Invalid Request, must follow JSON-RPC Request schema"))
			continue
		}

		s.dispatcher.Dispatch(request, connection)

		if request.Method == "chat" {
			var chatParams ChatRequestParams
			err = json.Unmarshal(request.Params, &chatParams)

			if err != nil {
				log.Printf("Failed deserializing request params: [%s] \n", request.Id)
				continue
			}

			params, _ := json.Marshal(ChatMessageNotification{Msg: chatParams.Msg})
			notification := JsonRpcNotification{JsonRpc: JsonRpcVersion, Method: ChatNotificationRpcMethod, Params: params}
			s.dispatcher.SendNotification(notification, ConnectionsToWriters(connections, connection))
		}

		clear(buf)
	}
}

func ConnectionsToWriters(connections []*Connection, exclude *Connection) []io.Writer {
	writers := make([]io.Writer, 0)

	for _, c := range connections {
		if c.id != exclude.id {
			writers = append(writers, c)
		}
	}

	return writers
}

func ChatMessageHandler(request JsonRpcRequest) JsonRpcResponse {
	var params ChatRequestParams
	err := json.Unmarshal(request.Params, &params)
	if err != nil {
		errorResponse := JsonRpcResponse{Id: request.Id, JsonRpc: request.JsonRpc, Error: &JsonRpcError{Code: -32600, Message: "Invalid Request"}}
		return errorResponse
	}

	result := SuccessResult{Success: true}
	resultJson, err := json.Marshal(result)
	response := JsonRpcResponse{Id: request.Id, JsonRpc: request.JsonRpc, Result: resultJson}
	return response
}

func main() {
	dispatcher := NewDispatcher()
	dispatcher.AddMethod("chat", ChatMessageHandler)
	connectionService := &ConnectionService{store: NewConnectionStore()}
	server := &Server{port: 8080, connectionService: connectionService, dispatcher: dispatcher}
	server.Start()
}
