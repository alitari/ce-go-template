package cetransformer

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/cloudevents/sdk-go/v2/protocol"
	"github.com/google/go-cmp/cmp"
)

// NewEventWithJSONStringData jsonDataString, source,type,id
func NewEventWithJSONStringData(jsonData string, st ...string) cloudevents.Event {
	var payload map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &payload)
	if err != nil {
		log.Fatal(err)
	}
	return NewEventWithMapData(payload, st...)
}

// NewEventWithMapData payload, source,type,id
func NewEventWithMapData(payload map[string]interface{}, st ...string) cloudevents.Event {
	event := createEventWithContext(st...)
	err := event.SetData(cloudevents.ApplicationJSON, payload)
	if err != nil {
		log.Fatal(err)
	}
	return event
}

// context = source,type,id
func createEventWithContext(context ...string) cloudevents.Event {
	event := cloudevents.NewEvent()
	if len(context) > 0 {
		event.SetSource(context[0])
	} else {
		event.SetSource("source")
	}
	if len(context) > 1 {
		event.SetType(context[1])
	} else {
		event.SetType("type")
	}
	if len(context) > 2 {
		event.SetID(context[2])
	} else {
		event.SetID("id")
	}
	event.SetDataContentType("application/json")
	event.SetSpecVersion("1.0")
	return event
}

// CompareEvents true if events are "equal"
func CompareEvents(t *testing.T, message string, actualEvent, wantedEvent cloudevents.Event) bool {
	if actualEvent.Source() != wantedEvent.Source() {
		t.Errorf("%s. source not fit.  actual= '%s', want '%v'", message, actualEvent.Source(), wantedEvent.Source())
		return false
	}

	if actualEvent.Type() != wantedEvent.Type() {
		t.Errorf("%s. type not fit.  actual= '%s', want '%s'", message, actualEvent.Type(), wantedEvent.Type())
		return false
	}
	if bytes.Compare(actualEvent.Data(), wantedEvent.Data()) != 0 {
		t.Errorf("%s. payload not equal. actual = '%v', want '%v'", message, actualEvent, wantedEvent)
		return false
	}
	return true
}

// CompareErrors true if events are "equal"
func CompareErrors(t *testing.T, message string, actualError, wantedError error) bool {
	if actualError == nil && wantedError == nil {
		return true
	}
	if actualError != nil && wantedError != nil {
		if actualError.Error() != wantedError.Error() {
			t.Errorf("%s unexpected error result, actual = '%v', want =  '%v'", message, actualError, wantedError)
			return false
		}
		return true
	}
	t.Errorf("%s unexpected error result, actual = '%v', want =  '%v'", message, actualError, wantedError)
	return false
}

// CeClientMock mock ce client
type CeClientMock struct {
	T                       *testing.T
	WantSend                bool
	WantSendEvent           cloudevents.Event
	ShouldThrowErrorOnStart error
	ShouldThrowErrorOnSend  error
}

// StartReceiver bla
func (mm *CeClientMock) StartReceiver(ctx context.Context, fn interface{}) error {
	mm.T.Logf("Callback func : %v", fn)
	return mm.ShouldThrowErrorOnStart
}

// Send bla
func (mm *CeClientMock) Send(ctx context.Context, event cloudevents.Event) protocol.Result {
	if !mm.WantSend {
		mm.T.Errorf("CeClientMock, Send should not be called: wantSend: %v", mm.WantSend)
	}
	if !cmp.Equal(mm.WantSendEvent, event) {
		mm.T.Errorf("CeClientMock, unexpected sourceEvent: actual: %v, but want %v", event, mm.WantSendEvent)
	}
	return mm.ShouldThrowErrorOnSend
}

// Request bla
func (mm *CeClientMock) Request(ctx context.Context, event cloudevents.Event) (*cloudevents.Event, protocol.Result) {
	mm.T.Errorf("CeClientMock, unexpected call: 'Request'")
	return nil, nil
}

// NewReq bla
func NewReq(method string, header http.Header, url, body string) *http.Request {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	req.Header = header
	if err != nil {
		log.Fatalf("Can't create request error = %v", err)
		return nil
	}
	return req
}

// NewGETRequest bla
func NewGETRequest(url string) *http.Request {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatalf("Can't create request error = %v", err)
		return nil
	}
	return req
}

// CompareResponses true if events are "equal"
func CompareResponses(t *testing.T, message string, actualResponse, wantedResponse http.Response) bool {
	if actualResponse.Status != wantedResponse.Status {
		t.Errorf("%s unexpected response status , actual = '%v', want =  '%v'", message, actualResponse.Status, wantedResponse.Status)
		return false
	}
	return true
}
