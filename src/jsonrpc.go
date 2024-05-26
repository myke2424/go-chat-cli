package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
)

const (
	JsonRpcVersion          = "2.0"
	ChatRpcMethod           = "chat"
	CreateUserRpcMethod     = "createUser"
	CreateChatRoomRpcMethod = "createChatRoom"
	DeleteChatRoomRpcMethod = "deleteChatRoom"
	JoinChatRoomRpcMethod   = "joinChatRoom"
	LeaveChatRoomRpcMethod  = "leaveChatRoom"
)

type JsonRpcRequest struct {
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []byte `json:"params"`
	Id      string `json:"id"`
}

func (r JsonRpcRequest) String() string {
	return fmt.Sprintf("JsonRpcRequest(id=%s, method=%s)", r.Id, r.Method)
}

type JsonRpcNotification struct {
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  []byte `json:"params"`
}

func (n JsonRpcNotification) String() string {
	return fmt.Sprintf("JsonRpcNotification(method=%s)", n.Method)
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
	Data    any    `json:"data,omitempty"`
}

func (e *JsonRpcError) String() string {
	return fmt.Sprintf("JsonRpcError(Code=%d, Message=%s", e.Code, e.Message)
}

type ChatRequestParams struct {
	Msg []byte `json:"msg"`
}

type SuccessResult struct {
	Success bool `json:"success"`
}

type ChatMessageNotification struct {
	Msg []byte `json:"msg"`
}

// JSON-RPC Handler function type
type RequestHandler func(request JsonRpcRequest) JsonRpcResponse

// JSON-RPC Request dispatcher for handling requests
type JsonRpcDispatcher struct {
	handlers map[string]RequestHandler
}

// Register a JSON-RPC Method with an associated handler func
func (d *JsonRpcDispatcher) AddMethod(method string, handler RequestHandler) {
	log.Printf("Adding rpc method: [%s]", method)
	d.handlers[method] = handler
}

// Invoke the handler if it exists
func (d *JsonRpcDispatcher) invokeHandler(request JsonRpcRequest) JsonRpcResponse {
	handler, ok := d.handlers[request.Method]
	if !ok {
		log.Printf("RPC Method not supported [%s]\n", request.Method)
		rpcError := &JsonRpcError{Code: -32601, Message: "Method not found"}
		return JsonRpcResponse{JsonRpc: request.JsonRpc, Id: request.Id, Error: rpcError}
	}
	return handler(request)
}

// Main interface for handling a JSON-RPC request and sending the response back to the client
func (d *JsonRpcDispatcher) Dispatch(request JsonRpcRequest, receiver io.Writer) error {
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
	_, err = receiver.Write(responseJson)
	if err != nil {
		log.Println("Failed to send json-rpc response to client")
		return err
	}

	log.Println("Response sent")
	return nil
}

// Send the JSON-RPC Notification to all the receivers
func (d *JsonRpcDispatcher) SendNotification(notification JsonRpcNotification, receivers []io.Writer) error {
	notificationJson, err := json.Marshal(notification)
	if err != nil {
		log.Println("Failed to seriailize json-rpc notification")
		return err
	}

	log.Printf("Broadcasting notification: [%s] to [%d] receivers\n", notification, len(receivers))
	for _, r := range receivers {
		_, err := r.Write(notificationJson)
		if err != nil {
			fmt.Println("Failed to send notification to receiver, moving on")
		}
	}

	return nil
}

// Initialise a new dispatcher
func NewDispatcher() *JsonRpcDispatcher {
	handlers := make(map[string]RequestHandler)
	dispatcher := &JsonRpcDispatcher{handlers: handlers}
	return dispatcher
}
