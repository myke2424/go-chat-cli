.PHONY: server client

# Command to start the server
server:
	go run server.go jsonrpc.go connection_service.go

# Command to start the client
client:
	go run client.go jsonrpc.go

