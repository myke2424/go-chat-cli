package main

import "fmt"
import "net"

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
		fmt.Println("Error starting server: ", err)
		panic("Failed to start server! Exiting")
	}

	s.listener = listener
	fmt.Printf("Server starting... Listening on port: [%d] \n", s.port)
	s.ListenForConnections()
}

func (s *Server) ListenForConnections() {
	for {
		fmt.Println("Accepting Connection...")
		conn, err := s.listener.Accept()
		if err != nil {
			fmt.Println("Error accepting incoming connection:", err)
			continue
		}
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
	fmt.Println("Adding new connection")
	c.connections = append(c.connections, connection)
	fmt.Println("Connection Added")
}

func (c *ConnectionManager) RemoveConnection(connection net.Conn) {
	fmt.Println("Removing connection")
	newConnections := make([]net.Conn, 0)

	for _, conn := range c.connections {
		if conn != connection {
			newConnections = append(newConnections, conn)
		}
	}

	c.connections = newConnections
	fmt.Println("Connection removed")
}

func (c *ConnectionManager) Send(message []byte, conn net.Conn) {
	fmt.Printf("Sending message: [%s]\n", string(message))
	_, err := conn.Write(message)

	if err != nil {
		fmt.Println("Error sending message: ", err)
		return
	}
	fmt.Println("Sent message successfully")
}

// Broadcast a message to all connections - excluding the connection who sent the message
func (c *ConnectionManager) Broadcast(message []byte, sender net.Conn) {
	receivers := make([]net.Conn, 0)
	for _, conn := range c.connections {
		if conn != sender {
			receivers = append(receivers, conn)
		}
	}

	fmt.Println("Broadcasting message to all connections: [%s]", string(message))
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
		fmt.Println("New connection")
		c.AddConnection(conn)
	} else {
		fmt.Println("Client is already connected...")
	}
	buf := make([]byte, 1024)

	bytesRead, err := conn.Read(buf)

	if err != nil {
		fmt.Println("Error reading buffer", err.Error())
		return
	}

	fmt.Printf("Read [%d] bytes from buffer\n", bytesRead)
	fmt.Printf("Received: %s\n", buf)
	c.Broadcast(buf, conn)
}

func (c *ConnectionManager) HandleDisconnect(conn net.Conn) {
	fmt.Println("Handling client disconnect")
	c.RemoveConnection(conn)
}

func main() {
	fmt.Println("Chatting..")
	cManager := &ConnectionManager{}
	cManager.Init()
	server := &Server{port: 8080, connectionManager: cManager, listener: nil}
	server.Start()
}
