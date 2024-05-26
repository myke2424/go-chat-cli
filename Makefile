.PHONY: server client

# Command to start the server
server:
	go run src/server.go src/jsonrpc.go src/connection_service.go

# Command to start the client
client:
	go run src/client.go src/jsonrpc.go

# Run tests
test:
	go test src/jsonrpc.go src/jsonrpc_test.go src/connection_service.go src/connection_service_test.go

