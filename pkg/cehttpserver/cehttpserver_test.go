package cehttpserver

import (
	"net/http"
	"testing"
	"time"

	"github.com/alitari/ce-go-template/pkg/cehandler"
	"github.com/alitari/ce-go-template/pkg/cetransformer"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
)

var input = "input"

var client = &http.Client{}

type CeProducerMock struct {
	t             *testing.T
	wantInputPath string
	wantInputBody string
	outgoingEvent cloudevents.Event
	shouldThrow   error
}

func (pm *CeProducerMock) CreateEvent(input interface{}) (*cloudevents.Event, error) {
	inputReq := input.(http.Request)
	if inputReq.URL.Path != pm.wantInputPath {
		pm.t.Errorf("CeProducerMock, unexpected request path: actual: %v, but want %v", inputReq.URL.Path, pm.wantInputPath)
	}
	return &pm.outgoingEvent, pm.shouldThrow
}

func TestCeHTTPServer_Serve(t *testing.T) {
	tests := []struct {
		name                   string
		givenProducerEvent     cloudevents.Event
		givenProducerError     protocol.Result
		givenCeClientSendError error
		givenServerPort        int
		givenServerMethod      string
		givenServerPath        string
		whenHTTPRequest        http.Request
		thenWantInputPath      string
		thenWantHTTPResponse   *http.Response
	}{
		{
			name:                 "Simple",
			givenServerMethod:    "GET",
			givenServerPath:      "/path",
			givenServerPort:      8088,
			givenProducerEvent:   cetransformer.NewEventWithJSONStringData(`{ "foo": "bar" }`),
			whenHTTPRequest:      *cetransformer.NewGETRequest("http://localhost:8088/path"),
			thenWantInputPath:    "/path",
			thenWantHTTPResponse: &http.Response{Status: "200 OK"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ceProducer := &CeProducerMock{t: t, wantInputPath: tt.thenWantInputPath, outgoingEvent: tt.givenProducerEvent, shouldThrow: tt.givenProducerError}
			ceClient := &cetransformer.CeClientMock{T: t, WantSend: true, WantSendEvent: tt.givenProducerEvent, ShouldThrowErrorOnSend: tt.givenCeClientSendError}
			ceProducerHandler := cehandler.NewProducerHandler(ceProducer, ceClient, "sink", 3*time.Second, true)
			ceHTTPServer := NewCeHTTPServer(tt.givenServerPort, tt.givenServerPath, tt.givenServerMethod, true, ceProducerHandler)
			defer ceHTTPServer.ShutDown()
			time.Sleep(100 * time.Millisecond)
			response, err := client.Do(&tt.whenHTTPRequest)
			if err != nil {
				t.Errorf("Couldn't send request= '%v', error: %v", tt.whenHTTPRequest, err)
			}
			cetransformer.CompareResponses(t, "CeHTTPServer", *response, *tt.thenWantHTTPResponse)
		})
	}
}
