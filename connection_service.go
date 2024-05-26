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
	Count() int
}

// Store connections in memory with a map
type InMemoryConnectionStore struct {
	connections map[string]*Connection
	count       int
}

// Create a new in-memory connection store
func NewConnectionStore() *InMemoryConnectionStore {
	connections := make(map[string]*Connection)
	return &InMemoryConnectionStore{connections: connections}
}

// Get the number of connections
func (s *InMemoryConnectionStore) Count() int {
	return s.count
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
	s.count--
}

// Insert a connection into the map
func (s *InMemoryConnectionStore) Add(connection *Connection) {
	s.connections[connection.id] = connection
	s.count++
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
