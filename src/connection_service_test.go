package main

import (
	"net"
	"strconv"
	"testing"
	"time"
)

const ConnectionCount = 20

type FakeNetConn struct {
	messagesWrote []string
}

func (f *FakeNetConn) Read(b []byte) (n int, err error) {
	messageCount := len(f.messagesWrote)
	if messageCount == 0 {
		return 0, nil
	}

	lastWrote := f.messagesWrote[messageCount-1]
	copy(b, []byte(lastWrote))
	return len(lastWrote), nil
}

func (f *FakeNetConn) Write(b []byte) (n int, err error) {
	f.messagesWrote = append(f.messagesWrote, string(b))
	return len(b), nil
}

func (f *FakeNetConn) Close() error {
	return nil
}

func (f *FakeNetConn) LocalAddr() net.Addr {
	return nil
}

func (f *FakeNetConn) RemoteAddr() net.Addr {
	return nil
}

func (f *FakeNetConn) SetDeadline(t time.Time) error {
	return nil
}

func (f *FakeNetConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (f *FakeNetConn) SetWriteDeadline(t time.Time) error {
	return nil
}

func NewFakeNetConn() *FakeNetConn {
	return &FakeNetConn{messagesWrote: make([]string, 0)}
}

// Connection store fixture with 20 connections
func ConnectionStoreFixture() *InMemoryConnectionStore {
	store := NewConnectionStore()

	for i := 0; i < ConnectionCount; i++ {
		connection := Connection{id: strconv.Itoa(i), Conn: NewFakeNetConn()}
		store.Add(&connection)
	}

	return store
}

func AssertNumberOfConnections(t testing.TB, got int, want int) {
	t.Helper()
	if got != want {
		t.Errorf("got [%d] but want [%d]", got, want)
	}
}

func AssertErrorNotNil(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Error("got error but wanted nil", err)
	}
}

func TestAddConnection(t *testing.T) {
	t.Run("add connection", func(t *testing.T) {
		service := &ConnectionService{store: ConnectionStoreFixture()}
		connection := &Connection{id: "21", Conn: NewFakeNetConn()}
		service.AddConnection(connection)

		got := service.store.Count()
		want := ConnectionCount + 1
		AssertNumberOfConnections(t, got, want)
	})
}

func TestDeleteConnection(t *testing.T) {
	t.Run("delete connection", func(t *testing.T) {
		service := &ConnectionService{store: ConnectionStoreFixture()}
		service.DeleteConnection("1")

		got := service.store.Count()
		want := ConnectionCount - 1
		AssertNumberOfConnections(t, got, want)

	})
}

func TestListConnections(t *testing.T) {
	t.Run("list connection", func(t *testing.T) {
		service := &ConnectionService{store: ConnectionStoreFixture()}
		connections := service.ListConnections()

		got := len(connections)
		want := ConnectionCount
		AssertNumberOfConnections(t, got, want)

	})
}

func TestGetConnection(t *testing.T) {
	t.Run("get existing connection", func(t *testing.T) {
		service := &ConnectionService{store: ConnectionStoreFixture()}
		connection, ok := service.store.Get("1")

		if !ok && connection == nil {
			t.Error("got nil but wanted connection")
		}

	})

	t.Run("get connection that doesnt exist", func(t *testing.T) {
		service := &ConnectionService{store: ConnectionStoreFixture()}
		connection, ok := service.store.Get("100")

		if ok && connection != nil {
			t.Error("got connection but wanted nil")
		}

	})
}
