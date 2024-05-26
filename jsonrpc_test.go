package main

import (
	"encoding/json"
	"io"
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

func DispatcherFixture() *JsonRpcDispatcher {
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

func (f *FakeWriter) AssertMessageReceived(t testing.TB, message string) {
	t.Helper()

	if len(f.data) == 0 {
		t.Error("No message receieved")
	}
	got := f.data[len(f.data)-1]

	if got != message {
		t.Errorf("Got message [%s] but wanted [%s]", got, message)
	}
}

func TestRpcDispatch(t *testing.T) {
	t.Run("request is processed and response is delivered to the destination", func(t *testing.T) {
		dispatcher := DispatcherFixture()
		fakeWriter := FakeWriter{data: make([]string, 0)}

		request := JsonRpcRequest{Id: "123", JsonRpc: JsonRpcVersion, Method: "add", Params: AddRequestParams{X: 5, Y: 10}}
		dispatcher.Dispatch(request, &fakeWriter)

		expectedResponse := `{"jsonrpc":"2.0","result":{"sum":15},"id":"123"}`
		fakeWriter.AssertMessageReceived(t, expectedResponse)
	})

	t.Run("unsupported rpc method returns error response", func(t *testing.T) {
		dispatcher := DispatcherFixture()
		fakeWriter := FakeWriter{data: make([]string, 0)}

		request := JsonRpcRequest{Id: "123", JsonRpc: JsonRpcVersion, Method: "fooBar"}
		dispatcher.Dispatch(request, &fakeWriter)

		expectedResponse := `{"jsonrpc":"2.0","error":{"code":-32601,"message":"Method not found"},"id":"123"}`
		fakeWriter.AssertMessageReceived(t, expectedResponse)
	})

	t.Run("notification is broadcasted to all receivers", func(t *testing.T) {
		dispatcher := DispatcherFixture()
		fakeWriters := make([]io.Writer, 0)

		for i := 0; i < 10; i++ {
			fakeWriter := &FakeWriter{data: make([]string, 0)}
			fakeWriters = append(fakeWriters, fakeWriter)
		}

		notification := JsonRpcNotification{JsonRpc: JsonRpcVersion, Method: "hello"}
		dispatcher.SendNotification(notification, fakeWriters)

		expectedNotification := `{"jsonrpc":"2.0","method":"hello","params":null}`
		for _, writer := range fakeWriters {
			fakeWriter := writer.(*FakeWriter)
			fakeWriter.AssertMessageReceived(t, expectedNotification)
		}

	})
}
