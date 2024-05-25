package main

import "fmt"

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
	Params  any    `json:"params"`
	Id      string `json:"id"`
}

func (r JsonRpcRequest) String() string {
	return fmt.Sprintf("JsonRpcRequest(id=%s, method=%s, params=%s", r.Id, r.Method, r.Params)
}

type JsonRpcNotification struct {
	JsonRpc string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  any    `json:"params"`
}

func (n JsonRpcNotification) String() string {
	return fmt.Sprintf("JsonRpcNotification(method=%s, params=%s", n.Method, n.Params)
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

type ChatRequestParams struct {
	Msg []byte `json:"msg"`
}
