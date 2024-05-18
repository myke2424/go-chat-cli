package main

import (
	"fmt"
	"log"
	"net"
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
		log.Println("Connection Accepted")
		go s.connectionManager.HandleConnect(conn)
	}
}

type ConnectionManager struct {
	connections []net.Conn
}

func (c *ConnectionManager) Init() {
	c.connections = make([]net.Conn, 0)
}

func (c *ConnectionManager) AddConnection(connection net.Conn) {
	c.connections = append(c.connections, connection)
	log.Println("New connection Added")
}

func (c *ConnectionManager) RemoveConnection(connection net.Conn) {
	log.Println("Removing connection")
	newConnections := make([]net.Conn, 0)

	for _, conn := range c.connections {
		if conn != connection {
			newConnections = append(newConnections, conn)
		}
	}

	c.connections = newConnections
	log.Println("Connection removed")
}

func (c *ConnectionManager) Send(message []byte, conn net.Conn) {
	log.Printf("Sending message: [%s]\n", string(message))
	_, err := conn.Write(message)

	if err != nil {
		log.Println("Error sending message: ", err)
		return
	}
	log.Println("Sent message successfully")
}

// Broadcast a message to all connections - excluding the connection who sent the message
func (c *ConnectionManager) Broadcast(message []byte, sender net.Conn) {
	receivers := make([]net.Conn, 0)
	for _, conn := range c.connections {
		if conn != sender {
			receivers = append(receivers, conn)
		}
	}

	log.Printf("Broadcasting message to all connections: [%s] \n", string(message))
	for _, conn := range receivers {
		c.Send(message, conn)
	}
}

func (c *ConnectionManager) HasConnection(conn net.Conn) bool {
	hasConnection := false
	for _, conn_ := range c.connections {
		if conn == conn_ {
			hasConnection = true
		}
	}
	return hasConnection
}

func (c *ConnectionManager) HandleConnect(conn net.Conn) {
	if !c.HasConnection(conn) {
		c.AddConnection(conn)
	}

	// Should I be doing this on a loop? Clear buffer iteration
	buf := make([]byte, 1024)

	for {
		bytesRead, err := conn.Read(buf) // Does 0 rv mean conn closed???

		if err != nil {
			// TODO: Catch error for large buffer size and push message to client msg is 2 big
			log.Fatalln("Error reading buffer", err.Error())
			return
		}

		log.Printf("Read [%d] bytes from buffer\n", bytesRead)
		log.Printf("Received: %s\n", buf)
		c.Broadcast(buf, conn)
		clear(buf)
	}
}

func (c *ConnectionManager) HandleDisconnect(conn net.Conn) {
	log.Println("Handling client disconnect")
	c.RemoveConnection(conn)
}

func main() {
	cManager := &ConnectionManager{}
	cManager.Init()
	server := &Server{port: 8080, connectionManager: cManager, listener: nil}
	server.Start()
}
