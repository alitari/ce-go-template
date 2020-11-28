package cetransformer

import (
	"bytes"
	"encoding/json"
	"log"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
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
