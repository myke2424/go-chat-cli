package main

import (
	"encoding/json"
	"io"
	"log"
)

// JSON-RPC Handler function type
type RequestHandler func(request JsonRpcRequest) JsonRpcResponse

// JSON-RPC Request dispatcher for handling requests
type JsonRpcRequestDispatcher struct {
	handlers map[string]RequestHandler
}

// Register a JSON-RPC Method with an associated handler func
func (d *JsonRpcRequestDispatcher) AddMethod(method string, handler RequestHandler) {
	log.Printf("Adding rpc method: [%s]", method)
	d.handlers[method] = handler
}

// Invoke the handler if it exists
func (d *JsonRpcRequestDispatcher) invokeHandler(request JsonRpcRequest) JsonRpcResponse {
	handler, ok := d.handlers[request.Method]
	if !ok {
		log.Printf("RPC Method not supported [%s]\n", request.Method)
		rpcError := &JsonRpcError{Code: -32601, Message: "Method not found"}
		return JsonRpcResponse{JsonRpc: request.JsonRpc, Id: request.Id, Error: rpcError}
	}
	return handler(request)
}

// Main interface for handling a JSON-RPC request and sending the response back to the client
func (d *JsonRpcRequestDispatcher) Dispatch(request JsonRpcRequest, destination io.Writer) error {
	response := d.invokeHandler(request)
	responseJson, err := json.Marshal(response)

	if err != nil {
		log.Println("Failed to serialize json-rpc response to JSON")
		errorResponse := JsonRpcResponse{Id: request.Id, JsonRpc: request.JsonRpc, Error: &JsonRpcError{Code: -32700, Message: "Parse error"}}
		responseJson, err = json.Marshal(errorResponse)
		if err != nil {
			return err
		}
	}

	log.Println("Sending json-rpc response")
	_, err = destination.Write(responseJson)
	if err != nil {
		log.Println("Failed to send json-rpc response to client")
		return err
	}

	log.Println("Response sent")
	return nil
}

// Initialise a new dispatcher
func NewDispatcher() *JsonRpcRequestDispatcher {
	handlers := make(map[string]RequestHandler)
	dispatcher := &JsonRpcRequestDispatcher{handlers: handlers}
	return dispatcher
}
