package cehttpclienttransformer

import (
	"encoding/json"
	"log"
	"reflect"
	"testing"
	"time"

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
		name         string
		se           cloudevents.Event
		de           cloudevents.Event
		ceTemplate   string
		httpTemplate string
		onlyPayload  bool
		wantErr      bool
	}{
		{name: "gender-male",
			httpTemplate: ReadFromFile("genderGet.http"),
			ceTemplate:   `{ "name": {{ .inputce.data.name | quote }}, "gender": {{ .httpresponse.body.gender | quote }} }`,
			onlyPayload:  false,
			se:           newEventJSON(`{"name": "Alex"}`),
			de:           newEventJSON(`{"name": "Alex", "gender": "male"}`),
			wantErr:      false},
		{name: "gender-female",
			httpTemplate: ReadFromFile("genderGet.http"),
			ceTemplate:   `{ "name": {{ .inputce.data.name | quote }}, "gender": {{ .httpresponse.body.gender | quote }} }`,
			onlyPayload:  false,
			se:           newEventJSON(`{"name": "Caroline"}`),
			de:           newEventJSON(`{"name": "Caroline", "gender": "female"}`),
			wantErr:      false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, err := NewCeHTTPClientTransformer(tt.httpTemplate, tt.ceTemplate, 5*time.Second, true, true, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("cehttpclienttransformer.TransformEvent error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := ct.TransformEvent(&tt.se)
			if (err != nil) != tt.wantErr {
				t.Errorf("cehttpclienttransformer.TransformEvent error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			want := tt.de
			gotData := map[string]interface{}{}
			got.DataAs(&gotData)
			wantData := map[string]interface{}{}
			want.DataAs(&wantData)
			if !reflect.DeepEqual(gotData, wantData) {
				t.Errorf("cehttpclienttransformer.TransformEvent data not equal:\nactual = '%v'\nwant   = '%v'", string(got.Data()), string(want.Data()))
			}
			if got.Source() != want.Source() {
				t.Errorf("cehttpclienttransformer.TransformEvent source not equal: actual = '%v', want '%v'", got.Source(), want.Source())
			}
			if got.Type() != want.Type() {
				t.Errorf("cehttpclienttransformer.TransformEvent type not equal: actual = %v, want %v", got.Type(), want.Type())
			}
		})
	}
}

func TestPredicateEvent(t *testing.T) {
	tests := []struct {
		name         string
		se           cloudevents.Event
		wanted       bool
		ceTemplate   string
		httpTemplate string
		wantErr      bool
	}{
		{name: "peter is not female",
			httpTemplate: ReadFromFile("genderGet.http"),
			ceTemplate:   `{{ eq .httpresponse.body.gender "female" | toString }}`,
			se:           newEventJSON(`{"name": "Peter"}`),
			wanted:       false,
			wantErr:      false},
		{name: "maria is female",
			httpTemplate: ReadFromFile("genderGet.http"),
			ceTemplate:   `{{ eq .httpresponse.body.gender "female" | toString }}`,
			se:           newEventJSON(`{"name": "Maria"}`),
			wanted:       true,
			wantErr:      false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ct, err := NewCeHTTPClientTransformer(tt.httpTemplate, tt.ceTemplate, 5*time.Second, true, true, true)
			if (err != nil) != tt.wantErr {
				t.Errorf("cehttpclienttransformer.PredicateEvent error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			got, err := ct.PredicateEvent(&tt.se)
			if (err != nil) != tt.wantErr {
				t.Errorf("cehttpclienttransformer.PredicateEvent error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wanted != got {
				t.Errorf("cehttpclienttransformer.PredicateEvent predicate not equal:  actual = '%v' , want = '%v'", got, tt.wanted)
			}
		})
	}
}

// func TestCreateEvent(t *testing.T) {
// 	tests := []struct {
// 		name        string
// 		request     *http.Request
// 		de          cloudevents.Event
// 		onlyPayload bool
// 		ceTemplate  string
// 		wantErr     bool
// 	}{
// 		{name: "simple",
// 			request:     nil,
// 			ceTemplate:  `{{ eq .httpresponse.body.gender "female" | toString }}`,
// 			de:          newEventJSON(`{"name": "Peter"}`),
// 			onlyPayload: true,
// 			wantErr:     false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ct := Newcehttpclienttransformer("", tt.ceTemplate, 5*time.Second, true, tt.onlyPayload, true)
// 			got, err := ct.CreateEvent(*tt.request)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("cehttpclienttransformer.CreateEvent error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			want := tt.de
// 			gotData := map[string]interface{}{}
// 			got.DataAs(&gotData)
// 			wantData := map[string]interface{}{}
// 			want.DataAs(&wantData)
// 			if !reflect.DeepEqual(gotData, wantData) {
// 				t.Errorf("cehttpclienttransformer.CreateEvent data not equal:\nactual = '%v'\nwant   = '%v'", string(got.Data()), string(want.Data()))
// 			}
// 			if got.Source() != want.Source() {
// 				t.Errorf("cehttpclienttransformer.CreateEvent source not equal: actual = '%v', want '%v'", got.Source(), want.Source())
// 			}
// 			if got.Type() != want.Type() {
// 				t.Errorf("cehttpclienttransformer.CreateEvent type not equal: actual = %v, want %v", got.Type(), want.Type())
// 			}
// 		})
// 	}
// }
