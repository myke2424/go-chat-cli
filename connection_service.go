package main

import (
	"log"
	"net"

	"github.com/google/uuid"
)

// Connection wrapped with an ID
type Connection struct {
	id string
	net.Conn
}

// Return the connectionId as the string
func (c *Connection) String() string {
	return c.id
}

// Connection Data store Interface
type ConnectionStore interface {
	Add(connection *Connection)
	Get(connectionId string) (*Connection, bool)
	List() []*Connection
	Delete(connectionId string)
}

// Store connections in memory with a map
type InMemoryConnectionStore struct {
	connections map[string]*Connection
}

// Get a connection by ID
func (s *InMemoryConnectionStore) Get(connectionId string) (*Connection, bool) {
	connection, ok := s.connections[connectionId]
	if !ok {
		return nil, false
	}
	return connection, true
}

// List all connections
func (s *InMemoryConnectionStore) List() []*Connection {
	connectionList := make([]*Connection, 0)

	for _, c := range s.connections {
		connectionList = append(connectionList, c)
	}

	return connectionList
}

// Remove a connection from the map
func (s *InMemoryConnectionStore) Delete(connectionId string) {
	delete(s.connections, connectionId)
}

// Insert a connection into the map
func (s *InMemoryConnectionStore) Add(connection *Connection) {
	s.connections[connection.id] = connection
}

// Connection Service for interfacing with connections
type ConnectionService struct {
	store ConnectionStore
}

// Add a new connection (wrapped net.Conn)
func (s *ConnectionService) AddConnection(conn net.Conn) {
	connectionId := uuid.New().String()
	log.Printf("Adding new connection: [%s]\n", connectionId)

	connection := Connection{id: connectionId, Conn: conn}
	s.store.Add(&connection)
}

// Delete a connection
func (s *ConnectionService) DeleteConnection(connectionId string) {
	log.Printf("Deleting connection: [%s]\n", connectionId)
	s.store.Delete(connectionId)
}

// List all connections
func (s *ConnectionService) ListConnections() []*Connection {
	log.Println("Listing all connections")
	return s.store.List()
}

// Get a connection
func (s *ConnectionService) GetConnection(connectionId string) (*Connection, bool) {
	log.Printf("Getting connection: [%s]\n", connectionId)
	return s.store.Get(connectionId)
}

// Send a message to the connection
func (s *ConnectionService) SendMessage(msg []byte, connection *Connection) {
	log.Printf("Sending message: [%s] to connection: [%s} \n", string(msg), connection.id)
	_, err := connection.Write(msg)

	if err != nil {
		log.Println("Error sending message: ", err)
		return
	}
	log.Println("Sent message successfully")

}

// Broadcast a message to all connections, with the option to exclude connections
func (s *ConnectionService) Broadcast(msg []byte, exclude []*Connection) {
	sliceToMapFn := func(connections []*Connection) map[string]*Connection {
		connectionMap := make(map[string]*Connection)
		for _, c := range connections {
			connectionMap[c.id] = c
		}
		return connectionMap
	}

	excludedConnections := sliceToMapFn(exclude)
	connections := s.store.List()

	if len(connections) == 0 {
		log.Println("No receivers, not broadcasting message.")
		return
	}

	log.Printf("Broadcasting message to all connections: [%s] \n", string(msg))
	for _, c := range connections {
		if _, ok := excludedConnections[c.id]; !ok {
			s.SendMessage(msg, c)
		}
	}
}
