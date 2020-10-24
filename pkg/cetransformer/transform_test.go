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

func TestCloudTransformer_Event_to_Event(t *testing.T) {
	tests := []struct {
		name    string
		se      cloudevents.Event
		de      cloudevents.Event
		config  CloudEventTransformerConfig
		wantErr bool
	}{
		{name: "constant",
			se:      newEventJSON(`{"name": "King"}`),
			config:  CloudEventTransformerConfig{CeTemplate: `{"data":{"name":"Alex"}, "datacontenttype":"application/json","id":"myid","source":"mysource","specversion":"1.0","type":"mytype"}`},
			de:      newEventJSON(`{"name": "Alex"}`, "mysource", "mytype", "myid"),
			wantErr: false},
		{name: "constantOnlyPayload",
			se:      newEventJSON(`{"name": "King"}`),
			config:  CloudEventTransformerConfig{CeTemplate: `{"name": "Alex"}`, OnlyPayload: true},
			de:      newEventJSON(`{"name": "Alex"}`),
			wantErr: false},
		{name: "base",
			se:      newEventJSON(`{"name": "King"}`, "mySource", "myType", "myId"),
			config:  CloudEventTransformerConfig{CeTemplate: `{"data": {"name":"{{ .data.name }}", "count":{{ count }} },"datacontenttype":"{{ .datacontenttype }}","id":"{{ .id }}","source":"{{ .source }}","specversion":"{{ .specversion }}","type":"{{ .type }}"}`},
			de:      newEventJSON(`{"count": 1, "name": "King" }`, "mySource", "myType", "myId"),
			wantErr: false},
		{name: "identOnlyPayload",
			se:      newEventJSON(`{"name": "King"}`),
			config:  CloudEventTransformerConfig{CeTemplate: `{{ toJson .data }}`, OnlyPayload: true},
			de:      newEventJSON(`{"name": "King" }`),
			wantErr: false},
		{name: "complex",
			se:      newEventJSON(`{"properties": { "key":"prop1", "value":"value1" }}`),
			config:  CloudEventTransformerConfig{CeTemplate: `{"data":{"{{.data.properties.key }}":"{{ .data.properties.value }}"},` + defaultEventJSONMetaData},
			de:      newEventJSON(`{"prop1": "value1"}`),
			wantErr: false},
		{name: "complexOnlyPayload",
			se:      newEventJSON(`{"properties": { "key":"prop1", "value":"value1" }}`),
			config:  CloudEventTransformerConfig{CeTemplate: `{"{{.data.properties.key }}":"{{ .data.properties.value }}"}`, OnlyPayload: true},
			de:      newEventJSON(`{"prop1": "value1"}`),
			wantErr: false},
		{name: "array",
			se:      newEventJSON(`{"values": [ "v1","v2"]}`),
			config:  CloudEventTransformerConfig{CeTemplate: `{ "data": {"value":"{{ index .data.values 0 }}"},` + defaultEventJSONMetaData},
			de:      newEventJSON(`{"value": "v1"}`),
			wantErr: false},
		{name: "sprig",
			se:      newEventJSON(`{"value": "Alex"}`),
			config:  CloudEventTransformerConfig{CeTemplate: `{ "data": {"value":"{{ b64enc .data.value }}"},` + defaultEventJSONMetaData},
			de:      newEventJSON(`{"value": "QWxleA=="}`),
			wantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.config.Debug = true
			ct := &CloudEventTransformer{
				Config: tt.config,
			}
			ct.Init()

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
