package main

import (
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
 *  Server JSON-RPC Methods.... [sendChatMsg, createUser, createChatRoom, joinChatRoom, leaveChatRoom]
 *                                req     req         req            req      req            req
 *
 *  Transport messages following JSON-RPC over TCP sockets...
 *
 */

type Server struct {
	port              int
	connectionManager *ConnectionManager
	listener          net.Listener
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", s.port))

	if err != nil {
		log.Fatalf("Failed to start server! Exiting", err)
	}

	s.listener = listener
	log.Printf("Server starting... Listening on port: [%d] \n", s.port)
	s.ListenForConnections()
}

func (s *Server) ListenForConnections() {
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

type Connection struct {
	id   string
	conn net.Conn
}

type ConnectionManager struct {
	connections map[string]Connection
}

func (c *ConnectionManager) Init() {
	c.connections = make(map[string]Connection)
}

func (c *ConnectionManager) AddConnection(connection Connection) {
	c.connections[connection.id] = connection
	log.Printf("New connection Added [%s}\n", connection.id)
}

func (c *ConnectionManager) RemoveConnection(connection Connection) {
	log.Printf("Removing connection: [%s]\n", connection.id)
	delete(c.connections, connection.id)
	log.Println("Connection removed")
}

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

func (c *ConnectionManager) HasConnection(conn Connection) bool {
	hasConnection := false
	for _, conn_ := range c.connections {
		if conn == conn_ {
			hasConnection = true
		}
	}
	return hasConnection
}

func (c *ConnectionManager) HandleConnect(connection Connection) {
	if !c.HasConnection(connection) {
		c.AddConnection(connection)
	}

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

func (c *ConnectionManager) HandleDisconnect(conn Connection) {
	log.Println("Handling client disconnect")
	c.RemoveConnection(conn)
}

func main() {
	cManager := &ConnectionManager{}
	cManager.Init()
	server := &Server{port: 8080, connectionManager: cManager, listener: nil}
	server.Start()
}
