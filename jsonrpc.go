package main

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
