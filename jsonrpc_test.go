package main

import (
	"encoding/json"
	"testing"
)

type AddRequestParams struct {
	X int `json:"x"`
	Y int `json:"y"`
}

type AddRequestResult struct {
	Sum int `json:"sum"`
}

func AddRequestHandler(request JsonRpcRequest) JsonRpcResponse {
	var params AddRequestParams
	paramsJson, _ := json.Marshal(request.Params)
	json.Unmarshal(paramsJson, &params)
	return JsonRpcResponse{Id: request.Id, JsonRpc: request.JsonRpc, Result: &AddRequestResult{Sum: params.X + params.Y}}
}

func DispatcherFixture() *JsonRpcRequestDispatcher {
	dispatcher := NewDispatcher()
	dispatcher.AddMethod("add", AddRequestHandler)
	return dispatcher
}

type FakeWriter struct {
	data []string
}

func (f *FakeWriter) Write(p []byte) (n int, err error) {
	f.data = append(f.data, string(p))
	return len(p), nil
}

func (f *FakeWriter) AssertResponseReceived(t testing.TB, expectedResponse string) {
	t.Helper()

	if len(f.data) == 0 {
		t.Error("No response receieved")
	}
	actualResponse := f.data[len(f.data)-1]

	if actualResponse != expectedResponse {
		t.Errorf("Got response [%s] but wanted [%s]", actualResponse, expectedResponse)
	}
}

func TestRpcDispatch(t *testing.T) {
	t.Run("request is processed and response is delivered to the destination", func(t *testing.T) {
		dispatcher := DispatcherFixture()
		fakeWriter := FakeWriter{}

		request := JsonRpcRequest{Id: "123", JsonRpc: JsonRpcVersion, Method: "add", Params: AddRequestParams{X: 5, Y: 10}}
		dispatcher.Dispatch(request, &fakeWriter)

		expectedResponse := `{"jsonrpc":"2.0","result":{"sum":15},"id":"123"}`
		fakeWriter.AssertResponseReceived(t, expectedResponse)
	})

	t.Run("unsupported rpc method returns error response", func(t *testing.T) {
		dispatcher := DispatcherFixture()
		fakeWriter := FakeWriter{}

		request := JsonRpcRequest{Id: "123", JsonRpc: JsonRpcVersion, Method: "fooBar"}
		dispatcher.Dispatch(request, &fakeWriter)

		expectedResponse := `{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"},"id":"123"}`
		fakeWriter.AssertResponseReceived(t, expectedResponse)
	})
}
