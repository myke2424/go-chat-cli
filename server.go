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

// TODO: Have some sort of config struct + yaml file to read in configuration as well

// TCP Server
type Server struct {
	port              int
	connectionManager *ConnectionManager
	listener          net.Listener
	dispatcher        *JsonRpcRequestDispatcher
}

// Connection wrapper which includes a unique ID
type Connection struct {
	id string
	net.Conn
}

func (c *Connection) Read(b []byte) (n int, err error) {
	return c.conn.Read(b)
}

func (c *Connection) Write(b []byte) (n int, err error) {
	return c.conn.Write(b)
}

// Struct for managing connection(s) state
type ConnectionManager struct {
	connections map[string]Connection
}

// JSON-RPC Handler function type
type RequestHandler func(request JsonRpcRequest) JsonRpcResponse

// JSON-RPC Request dispatcher for handling requests
type JsonRpcRequestDispatcher struct {
	handlers map[string]RequestHandler
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
		if err != nil {
			log.Println("Error accepting incoming connection:", err)
			continue
		}
		connectionId := uuid.New().String()
		log.Printf("Connection Accepted. ConnectionId = [%s]\n", connectionId)
		connection := Connection{id: connectionId, conn: conn}
		go s.connectionManager.HandleConnect(connection)
	}
}

// Initialize the connection manager
func (c *ConnectionManager) Init() {
	c.connections = make(map[string]Connection)
}

// Add a connection
func (c *ConnectionManager) AddConnection(connection Connection) {
	c.connections[connection.id] = connection
	log.Printf("New connection Added [%s}\n", connection.id)
}

// Remove a connection
func (c *ConnectionManager) RemoveConnection(connection Connection) {
	log.Printf("Removing connection: [%s]\n", connection.id)
	delete(c.connections, connection.id)
	log.Println("Connection removed")
}

// Send a message to a connection
func (c *ConnectionManager) Send(message []byte, connection Connection) {
	log.Printf("Sending message: [%s] to connection: [%s} \n", string(message), connection.id)
	_, err := connection.conn.Write(message)

	if err != nil {
		log.Println("Error sending message: ", err)
		return
	}
	log.Println("Sent message successfully")
}

// Broadcast a message to all connections - excluding the connection who sent the message
// TODO: Add exclude method for exlcuding slice of connections
func (c *ConnectionManager) Broadcast(message []byte, sender Connection) {
	receivers := make([]Connection, 0)
	for _, conn := range c.connections {
		if conn != sender {
			receivers = append(receivers, conn)
		}
	}

	if len(receivers) == 0 {
		log.Println("No receivers, not broadcasting message.")
		return
	}

	log.Printf("Broadcasting message to all connections: [%s] \n", string(message))
	for _, conn := range receivers {
		c.Send(message, conn)
	}
}

// Check if the connection exists
func (c *ConnectionManager) HasConnection(conn Connection) bool {
	hasConnection := false
	for _, conn_ := range c.connections {
		if conn == conn_ {
			hasConnection = true
		}
	}
	return hasConnection
}

// Handler for new connections
func (c *ConnectionManager) HandleConnect(connection Connection) {
	if !c.HasConnection(connection) {
		c.AddConnection(connection)
	}

	// Extract out reading from the connection socket into seperate go-routine.
	// HandleConnect should just add the connection to list of connections

	// Should I be doing this on a loop? Clear buffer iteration
	buf := make([]byte, 1024)

	for {
		bytesRead, err := connection.conn.Read(buf) // Does 0 rv mean conn closed???

		if err != nil {
			// TODO: Catch error for large buffer size and push message to client msg is 2 big
			log.Fatalln("Error reading connection buffer", err.Error())
			return
		}

		log.Printf("Read [%d] bytes from connection [%s} \n", bytesRead, connection.id)
		log.Printf("Message: %s\n", buf)
		c.Broadcast(buf, connection)
		clear(buf)
	}
}

// Handler for clients who have disconnected
func (c *ConnectionManager) HandleDisconnect(conn Connection) {
	log.Println("Handling client disconnect")
	c.RemoveConnection(conn)
}

func (d *JsonRpcRequestDispatcher) Init() {
	d.handlers = make(map[string]RequestHandler)
}

// Main interface for listening for messages
func (d *JsonRpcRequestDispatcher) ListenForMessages() {
	buf := make([]byte, 1024)
	for {
		bytesRead, err := connection.conn.Read(buf) // Does 0 rv mean conn closed???

		if err != nil {
			// TODO: Catch error for large buffer size and push message to client msg is 2 big
			log.Fatalln("Error reading connection buffer", err.Error())
			return
		}

		log.Printf("Read [%d] bytes from connection [%s} \n", bytesRead, connection.id)
		log.Printf("Message: %s\n", buf)
		c.Broadcast(buf, connection)
		clear(buf)
	}

}

// Register a JSON-RPC Method with an associated handler func
func (d *JsonRpcRequestDispatcher) AddMethod(method string, handler RequestHandler) {
	log.Printf("Adding rpc method: [%s]", method)
	d.handlers[method] = handler
}

// Invoke the handler if it exists
func (d *JsonRpcRequestDispatcher) InvokeHandler(request JsonRpcRequest) JsonRpcResponse {
	handler, ok := d.handlers[request.Method]
	if !ok {
		// TODO: Send rpc error back to client, method not found
		panic("rpc method not found")
	}
	return handler(request)
}

// Main interface for handling a JSON-RPC request and sending the response back to the client
func (d *JsonRpcRequestDispatcher) Dispatch(request JsonRpcRequest, conn Connection) {
	response := d.InvokeHandler(request)
	responseJson, err := json.Marshal(response)
	if err != nil {
		// TODO: If this fails... Send rpc internal server error to client
		panic("Failed to serialize rpc response")
	}

	log.Println("Sending json rpc response")
	_, err = conn.Write(responseJson)
	if err != nil {
		panic("Failed to send response")
	}
	log.Println("Response sent")

}

func main() {
	dispatcher := &JsonRpcRequestDispatcher{}
	dispatcher.Init()
	manager := &ConnectionManager{}
	manager.Init()

	server := &Server{port: 8080, connectionManager: manager, listener: nil, dispatcher: dispatcher}
	server.Start()
}
