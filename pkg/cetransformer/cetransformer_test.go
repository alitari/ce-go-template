package cetransformer

import (
	"math/rand"
	"reflect"
	"testing"

	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const defaultEventJSONMetaData = `"datacontenttype":"application/json","id":"id","source":"source","specversion":"1.0","type":"type"}`

func TestTransformEvent(t *testing.T) {
	tests := []struct {
		name             string
		givenTemplate    string
		givenOnlyCreate  bool
		givenEventType   string
		givenEventSource string
		whenEvent        cloudevents.Event
		thenEvent        cloudevents.Event
		thenError        bool
	}{

		{name: "template syntax error",
			givenTemplate: `{{ notexists }}`,
			whenEvent:     NewEventWithJSONStringData(`{"name": "King"}`),
			thenEvent:     cloudevents.NewEvent(),
			thenError:     true},
		{name: "template execution error",
			givenTemplate: `{{ .data.name | round }}`,
			whenEvent:     NewEventWithJSONStringData(`{"name": "King"}`),
			thenEvent:     cloudevents.NewEvent(),
			thenError:     true},
		{name: "template has wrong Json",
			givenTemplate: `{`,
			whenEvent:     NewEventWithJSONStringData(`{"name": "King"}`),
			thenEvent:     cloudevents.NewEvent(),
			thenError:     true},
		{name: "constant template",
			givenTemplate: `{"name": "Alex"}`,
			whenEvent:     NewEventWithJSONStringData(`{"name": "King"}`),
			thenEvent:     NewEventWithJSONStringData(`{"name": "Alex"}`),
			thenError:     false},
		{name: "constantCreate",
			givenOnlyCreate: true,
			givenTemplate:   `{"name": "Alex"}`,
			whenEvent:       cloudevents.NewEvent(),
			thenEvent:       NewEventWithJSONStringData(`{"name": "Alex"}`, "", ""),
			thenError:       false},
		{name: "constantType",
			givenTemplate:  `{"name": "Alex"}`,
			givenEventType: "danitype",
			whenEvent:      NewEventWithJSONStringData(`{"name": "King"}`),
			thenEvent:      NewEventWithJSONStringData(`{"name": "Alex"}`, "source", "danitype"),
			thenError:      false},
		{name: "constantSourceAndType",
			givenTemplate:    `{"name": "Alex"}`,
			givenEventType:   "alexType",
			givenEventSource: "alexSource",
			whenEvent:        NewEventWithJSONStringData(`{"name": "King"}`),
			thenEvent:        NewEventWithJSONStringData(`{"name": "Alex"}`, "alexSource", "alexType"),
			thenError:        false},

		{name: "identOnlyPayload",
			givenTemplate: `{{ toJson .data }}`,
			whenEvent:     NewEventWithJSONStringData(`{"name": "King"}`),
			thenEvent:     NewEventWithJSONStringData(`{"name": "King" }`),
			thenError:     false},

		{name: "complexOnlyPayload",
			givenTemplate: `{"{{.data.properties.key }}":"{{ .data.properties.value }}"}`,
			whenEvent:     NewEventWithJSONStringData(`{"properties": { "key":"prop1", "value":"value1" }}`),
			thenEvent:     NewEventWithJSONStringData(`{"prop1": "value1"}`),
			thenError:     false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var ct *CloudEventTransformer
			var err error
			if tt.givenEventType != "" {
				if tt.givenEventSource != "" {
					ct, err = NewCloudEventTransformer(tt.givenTemplate, tt.givenEventSource, tt.givenEventType, rand.Float32() < 0.5)
				} else {
					ct, err = NewCloudEventTransformer(tt.givenTemplate, tt.givenEventSource, tt.givenEventType, rand.Float32() < 0.5)
				}
			} else {
				ct, err = NewCloudEventTransformer(tt.givenTemplate, "", "", rand.Float32() < 0.5)
			}
			if err != nil {
				if !tt.thenError {
					t.Errorf("CloudEventTransformer.TransformEvent error = %v, wantErr %v", err, tt.thenError)
				}
				if ct != nil {
					t.Errorf("CloudEventTransformer must be nil, if error, but is %v", ct)
				}
				return
			}
			var got *cloudevents.Event
			if tt.givenOnlyCreate {
				got, err = ct.CreateEvent(nil)
			} else {
				got, err = ct.TransformEvent(&tt.whenEvent)
			}
			if err != nil {
				if !tt.thenError {
					t.Errorf("CloudEventTransformer.TransformEvent error = %v, wantErr %v", err, tt.thenError)
				}
				if got != nil {
					t.Errorf("CloudEventTransformer.TransformEvent() must be nil, if error, but is %v", got)
				}
				return
			}
			want := tt.thenEvent
			gotData := map[string]interface{}{}
			got.DataAs(&gotData)
			wantData := map[string]interface{}{}
			want.DataAs(&wantData)
			CompareEvents(t, "CloudEventTransformer.TransformEvent", *got, want)
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
			se:         NewEventWithJSONStringData(`{"name": "King"}`),
			ceTemplate: `true`,
			want:       true,
			wantErr:    false},

		{name: "constantFalse",
			se:         NewEventWithJSONStringData(`{"name": "King"}`),
			ceTemplate: ``,
			want:       false,
			wantErr:    false},

		{name: "simpleTrue",
			se:         NewEventWithJSONStringData(`{"name": "King"}`),
			ceTemplate: `{{ eq .data.name "King" | toString }}`,
			want:       true,
			wantErr:    false},

		{name: "simpleSourceFilter",
			se:         NewEventWithJSONStringData(`{"name": "King"}`, "mysource"),
			ceTemplate: `{{ eq .source "mysource" | toString }}`,
			want:       true,
			wantErr:    false},

		{name: "simpleWrong",
			se:         NewEventWithJSONStringData(`{"name": "King"}`),
			ceTemplate: `{{ eq .data.name "Queen" | toString }}`,
			want:       false,
			wantErr:    false},

		{name: "simpleError",
			se:         NewEventWithJSONStringData(`{"name": "King"}`),
			ceTemplate: `{{ eq .data.foo "Queen" | toString }}`,
			want:       false,
			wantErr:    true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, err := NewCloudEventTransformer(tt.ceTemplate, "", "", true)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("CloudEventTransformer.TransformEvent error = %v, wantErr %v", err, tt.wantErr)
				}
				if ct != nil {
					t.Errorf("CloudEventTransformer must be nil, if error, but is %v", ct)
				}
				return
			}
			got, err := ct.PredicateEvent(&tt.se)
			if err != nil {
				if !tt.wantErr {
					t.Errorf("CloudEventTransformer.TransformEvent error = %v, wantErr %v", err, tt.wantErr)
				}
				return
			}
			if got != tt.want {
				t.Errorf("CloudEventTransformer.PredicateEvent result not equal:\nactual = '%v'\nwant   = '%v'", got, tt.want)
			}
		})
	}
}

func TestUnmarshal(t *testing.T) {
	tests := []struct {
		name          string
		whenString    string
		thenWantEvent cloudevents.Event
		thenWantErr   bool
	}{
		{name: "empty",
			whenString:    "{}",
			thenWantEvent: NewEventWithMapData(map[string]interface{}{}, "", ""),
			thenWantErr:   false,
		},
		{name: "nojson",
			whenString:    "{",
			thenWantEvent: NewEventWithMapData(map[string]interface{}{}, "", ""),
			thenWantErr:   true,
		},
		{name: "simple",
			whenString:    `{ "name": "Alex"}`,
			thenWantEvent: NewEventWithMapData(map[string]interface{}{"name": "Alex"}, "", ""),
			thenWantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resultEvent := cloudevents.NewEvent()
			if err := Unmarshal([]byte(tt.whenString), &resultEvent); (err != nil) != tt.thenWantErr {
				t.Errorf("Unmarshal() error = %v, wantErr %v", err, tt.thenWantErr)
			}
			CompareEvents(t, "Unmarshal event", resultEvent, tt.thenWantEvent)
		})
	}
}

func TestEventToMap(t *testing.T) {
	tests := []struct {
		name        string
		whenEvent   cloudevents.Event
		thenWantMap map[string]interface{}
	}{
		{name: "empty",
			whenEvent:   NewEventWithJSONStringData("{}"),
			thenWantMap: map[string]interface{}{"data": map[string]interface{}{}, "source": "source", "type": "type", "datacontenttype": "application/json", "specversion": "1.0", "id": "id"},
		},
		{name: "emptyWith source and type and id",
			whenEvent:   NewEventWithJSONStringData("{}", "mysource", "mytype", "myid"),
			thenWantMap: map[string]interface{}{"data": map[string]interface{}{}, "source": "mysource", "type": "mytype", "datacontenttype": "application/json", "specversion": "1.0", "id": "myid"},
		},
		{name: "jsondata",
			whenEvent:   NewEventWithJSONStringData(`{ "foo": "bar"}`, "mysource", "mytype", "myid"),
			thenWantMap: map[string]interface{}{"data": map[string]interface{}{"foo": "bar"}, "source": "mysource", "type": "mytype", "datacontenttype": "application/json", "specversion": "1.0", "id": "myid"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualMap := EventToMap(&tt.whenEvent)
			if !reflect.DeepEqual(actualMap, tt.thenWantMap) {
				t.Errorf("actual map = '%v', want map '%v'", actualMap, tt.thenWantMap)
				return
			}
		})
	}
}

func TestCloudEventTransformer_TransformBytesToEvent(t *testing.T) {
	tests := []struct {
		name                  string
		whenEventDataAsString string
		whenEventContext      cloudevents.EventContext
		thenWantEvent         cloudevents.Event
		thenWantErr           bool
	}{
		{name: "simple",
			whenEventDataAsString: "{}",
			whenEventContext:      &cloudevents.EventContextV03{Type: "mytype"},
			thenWantErr:           false,
			thenWantEvent:         NewEventWithMapData(map[string]interface{}{}, "", "mytype"),
		},
		{name: "wrongjson",
			whenEventDataAsString: "}",
			whenEventContext:      &cloudevents.EventContextV03{Type: "mytype"},
			thenWantErr:           true,
			thenWantEvent:         NewEventWithMapData(map[string]interface{}{}, "", "mytype"),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actualEvent, err := TransformBytesToEvent([]byte(tt.whenEventDataAsString), tt.whenEventContext)
			if (err != nil) != tt.thenWantErr {
				t.Errorf("CloudEventTransformer.TransformBytesToEvent() error = %v, wantErr %v", err, tt.thenWantErr)
				return
			}
			if tt.thenWantErr {
				if actualEvent != nil {
					t.Errorf("CloudEventTransformer.TransformBytesToEvent() must be nil, if error, but is %v", actualEvent)
					return
				}
			} else {
				CompareEvents(t, "CloudEventTransformer.TransformBytesToEvent()", *actualEvent, tt.thenWantEvent)
			}
		})
	}
}
