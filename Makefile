.PHONY: server client

# Command to start the server
server:
	go run src/server.go src/jsonrpc.go src/connection_service.go

# Command to start the client
client:
	go run src/client.go src/jsonrpc.go

