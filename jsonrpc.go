package main

import (
	"encoding/json"
	"io"
	"log"
)

const JsonRpcVersionTwo = "2.0"

type JsonRpcRequest struct {
	JsonRpc string         `json:"jsonrpc"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
	Id      string         `json:"id"`
}

type JsonRpcNotification struct {
	JsonRpc string         `json:"jsonrpc"`
	Method  string         `json:"method"`
	Params  map[string]any `json:"params"`
}

type JsonRpcResponse struct {
	JsonRpc string        `json:"jsonrpc"`
	Result  any           `json:"result,omitempty"`
	Error   *JsonRpcError `json:"error,omitempty"`
	Id      string        `json:"id"`
}

type JsonRpcError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data"`
}

type RequestHandler func(request JsonRpcRequest) JsonRpcResponse

type RequestDispatcher struct {
	handlers  map[string]RequestHandler
	transport io.Writer
}

func (d *RequestDispatcher) Init() {
	d.handlers = make(map[string]RequestHandler)
}

func (d *RequestDispatcher) AddMethod(method string, handler RequestHandler) {
	log.Printf("Adding rpc method: [%s]", method)
	d.handlers[method] = handler
}

func (d *RequestDispatcher) invokeHandler(request JsonRpcRequest) JsonRpcResponse {
	handler, ok := d.handlers[request.Method]
	if !ok {
		// TODO: Send rpc error back to client, method not found
		panic("rpc method not found")
	}
	return handler(request)
}

func (d *RequestDispatcher) dispatch(request JsonRpcRequest) {
	response := d.invokeHandler(request)
	responseJson, err := json.Marshal(response)
	if err != nil {
		// TODO: If this fails... Send rpc internal server error to client
		panic("Failed to serialize rpc response")
	}

	log.Println("Sending json rpc response")
	_, err = d.transport.Write(responseJson)
	if err != nil {
		panic("Failed to send response")
	}
	log.Println("Response sent")
	// TODO: Send response back to client

}
