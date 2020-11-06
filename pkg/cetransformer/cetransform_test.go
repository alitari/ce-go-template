package cetransformer

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

// jsonData, source,type,id
func newEventJSON(jsonData string, st ...string) cloudevents.Event {
	event := cloudevents.NewEvent()
	if len(st) > 0 {
		event.SetSource(st[0])
	} else {
		event.SetSource("source")
	}
	if len(st) > 1 {
		event.SetType(st[1])
	} else {
		event.SetType("type")
	}
	if len(st) > 2 {
		event.SetID(st[2])
	} else {
		event.SetID("id")
	}
	event.SetDataContentType("application/json")
	event.SetSpecVersion("1.0")
	var payload map[string]interface{}
	err := json.Unmarshal([]byte(jsonData), &payload)
	if err != nil {
		log.Fatal(err)
	}
	err = event.SetData(cloudevents.ApplicationJSON, payload)
	if err != nil {
		log.Fatal(err)
	}
	return event
}

const defaultEventJSONMetaData = `"datacontenttype":"application/json","id":"id","source":"source","specversion":"1.0","type":"type"}`

func TestTransformEvent(t *testing.T) {
	tests := []struct {
		name        string
		se          cloudevents.Event
		de          cloudevents.Event
		ceTemplate  string
		onlyPayload bool
		wantErr     bool
	}{
		{name: "constant",
			ceTemplate:  `{"data":{"name":"Alex"}, "datacontenttype":"application/json","id":"myid","source":"mysource","specversion":"1.0","type":"mytype"}`,
			onlyPayload: false,
			se:          newEventJSON(`{"name": "King"}`),
			de:          newEventJSON(`{"name": "Alex"}`, "mysource", "mytype", "myid"),
			wantErr:     false},
		{name: "constantOnlyPayload",
			ceTemplate:  `{"name": "Alex"}`,
			onlyPayload: true,
			se:          newEventJSON(`{"name": "King"}`),
			de:          newEventJSON(`{"name": "Alex"}`),
			wantErr:     false},
		{name: "base",
			ceTemplate:  `{"data": {"name":"{{ .data.name }}", "count":{{ count }} },"datacontenttype":"{{ .datacontenttype }}","id":"{{ .id }}","source":"{{ .source }}","specversion":"{{ .specversion }}","type":"{{ .type }}"}`,
			onlyPayload: false,
			se:          newEventJSON(`{"name": "King"}`, "mySource", "myType", "myId"),
			de:          newEventJSON(`{"count": 1, "name": "King" }`, "mySource", "myType", "myId"),
			wantErr:     false},
		{name: "identOnlyPayload",
			ceTemplate:  `{{ toJson .data }}`,
			onlyPayload: true,
			se:          newEventJSON(`{"name": "King"}`),
			de:          newEventJSON(`{"name": "King" }`),
			wantErr:     false},
		{name: "complex",
			ceTemplate:  `{"data":{"{{.data.properties.key }}":"{{ .data.properties.value }}"},` + defaultEventJSONMetaData,
			onlyPayload: false,
			se:          newEventJSON(`{"properties": { "key":"prop1", "value":"value1" }}`),
			de:          newEventJSON(`{"prop1": "value1"}`),
			wantErr:     false},
		{name: "complexOnlyPayload",
			ceTemplate:  `{"{{.data.properties.key }}":"{{ .data.properties.value }}"}`,
			onlyPayload: true,
			se:          newEventJSON(`{"properties": { "key":"prop1", "value":"value1" }}`),
			de:          newEventJSON(`{"prop1": "value1"}`),
			wantErr:     false},
		{name: "array",
			ceTemplate:  `{ "data": {"value":"{{ index .data.values 0 }}"},` + defaultEventJSONMetaData,
			onlyPayload: false,
			se:          newEventJSON(`{"values": [ "v1","v2"]}`),
			de:          newEventJSON(`{"value": "v1"}`),
			wantErr:     false},
		{name: "sprig",
			ceTemplate:  `{ "data": {"value":"{{ b64enc .data.value }}"},` + defaultEventJSONMetaData,
			onlyPayload: false,
			se:          newEventJSON(`{"value": "Alex"}`),
			de:          newEventJSON(`{"value": "QWxleA=="}`),
			wantErr:     false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := NewCloudEventTransformer(tt.ceTemplate, tt.onlyPayload, true)
			got, err := ct.TransformEvent(&tt.se)
			if (err != nil) != tt.wantErr {
				t.Errorf("CloudEventTransformer.TransformEvent error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := tt.de
			gotData := map[string]interface{}{}
			got.DataAs(&gotData)
			wantData := map[string]interface{}{}
			want.DataAs(&wantData)
			if !reflect.DeepEqual(gotData, wantData) {
				t.Errorf("CloudEventTransformer.TransformEvent data not equal:\nactual = '%v'\nwant   = '%v'", string(got.Data()), string(want.Data()))
			}
			if got.Source() != want.Source() {
				t.Errorf("CloudEventTransformer.TransformEvent source not equal: actual = '%v', want '%v'", got.Source(), want.Source())
			}
			if got.Type() != want.Type() {
				t.Errorf("CloudEventTransformer.TransformEvent type not equal: actual = %v, want %v", got.Type(), want.Type())
			}
		})
	}
}

func TestPredicateEvent(t *testing.T) {
	tests := []struct {
		name       string
		se         cloudevents.Event
		want       bool
		ceTemplate string
		wantErr    bool
	}{
		{name: "constantTrue",
			se:         newEventJSON(`{"name": "King"}`),
			ceTemplate: `true`,
			want:       true,
			wantErr:    false},

		{name: "constantFalse",
			se:         newEventJSON(`{"name": "King"}`),
			ceTemplate: ``,
			want:       false,
			wantErr:    false},

		{name: "simpleTrue",
			se:         newEventJSON(`{"name": "King"}`),
			ceTemplate: `{{ eq .data.name "King" | toString }}`,
			want:       true,
			wantErr:    false},

		{name: "simpleSourceFilter",
			se:         newEventJSON(`{"name": "King"}`, "mysource"),
			ceTemplate: `{{ eq .source "mysource" | toString }}`,
			want:       true,
			wantErr:    false},

		{name: "simpleWrong",
			se:         newEventJSON(`{"name": "King"}`),
			ceTemplate: `{{ eq .data.name "Queen" | toString }}`,
			want:       false,
			wantErr:    false},

		{name: "simpleError",
			se:         newEventJSON(`{"name": "King"}`),
			ceTemplate: `{{ eq .data.foo "Queen" | toString }}`,
			want:       false,
			wantErr:    true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct := NewCloudEventTransformer(tt.ceTemplate, true, true)
			got, err := ct.PredicateEvent(&tt.se)
			if (err != nil) != tt.wantErr {
				t.Errorf("CloudEventTransformer.PredicateEvent error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if got != tt.want {
				t.Errorf("CloudEventTransformer.PredicateEvent result not equal:\nactual = '%v'\nwant   = '%v'", got, tt.want)
			}
		})
	}
}
