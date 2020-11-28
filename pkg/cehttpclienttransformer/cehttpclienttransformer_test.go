package cehttpclienttransformer

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/alitari/ce-go-template/pkg/cetransformer"
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

// func TestTransformEvent(t *testing.T) {
// 	tests := []struct {
// 		name         string
// 		se           cloudevents.Event
// 		de           cloudevents.Event
// 		ceTemplate   string
// 		httpTemplate string
// 		onlyPayload  bool
// 		wantErr      bool
// 	}{
// 		{name: "gender-male",
// 			httpTemplate: ReadFromFile("genderGet.http"),
// 			ceTemplate:   `{ "name": {{ .inputce.data.name | quote }}, "gender": {{ .httpresponse.body.gender | quote }} }`,
// 			onlyPayload:  false,
// 			se:           newEventJSON(`{"name": "Alex"}`),
// 			de:           newEventJSON(`{"name": "Alex", "gender": "male"}`),
// 			wantErr:      false},
// 		{name: "gender-female",
// 			httpTemplate: ReadFromFile("genderGet.http"),
// 			ceTemplate:   `{ "name": {{ .inputce.data.name | quote }}, "gender": {{ .httpresponse.body.gender | quote }} }`,
// 			onlyPayload:  false,
// 			se:           newEventJSON(`{"name": "Caroline"}`),
// 			de:           newEventJSON(`{"name": "Caroline", "gender": "female"}`),
// 			wantErr:      false},
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ct, err := NewCeHTTPClientTransformer(tt.httpTemplate, tt.ceTemplate, 5*time.Second, true, true, true)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("cehttpclienttransformer.TransformEvent error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			got, err := ct.TransformEvent(&tt.se)
// 			if (err != nil) != tt.wantErr {
// 				t.Errorf("cehttpclienttransformer.TransformEvent error = %v, wantErr %v", err, tt.wantErr)
// 				return
// 			}
// 			want := tt.de
// 			gotData := map[string]interface{}{}
// 			got.DataAs(&gotData)
// 			wantData := map[string]interface{}{}
// 			want.DataAs(&wantData)
// 			if !reflect.DeepEqual(gotData, wantData) {
// 				t.Errorf("cehttpclienttransformer.TransformEvent data not equal:\nactual = '%v'\nwant   = '%v'", string(got.Data()), string(want.Data()))
// 			}
// 			if got.Source() != want.Source() {
// 				t.Errorf("cehttpclienttransformer.TransformEvent source not equal: actual = '%v', want '%v'", got.Source(), want.Source())
// 			}
// 			if got.Type() != want.Type() {
// 				t.Errorf("cehttpclienttransformer.TransformEvent type not equal: actual = %v, want %v", got.Type(), want.Type())
// 			}
// 		})
// 	}
// }

type MockHTTPSender struct {
	thenResponse http.Response
}

func NewMockHTTPSender(thenResponse http.Response) *MockHTTPSender {
	ms := &MockHTTPSender{thenResponse: thenResponse}
	return ms
}

func (ms *MockHTTPSender) Send() (*http.Response, error) {
	return &ms.thenResponse, nil
}

func TestPredicateEvent(t *testing.T) {
	tests := []struct {
		name                  string
		givenHTTPTemplate     string
		givenCeTemplate       string
		givenHTTPResponse     http.Response
		whenIncomingEvent     cloudevents.Event
		thenWantHTTPProtocol  string
		thenWantOutgoingEvent cloudevents.Event
		thenWantErr           bool
	}{
		{name: "Get constant template",
			givenHTTPTemplate: `GET http://localhost:8080/get HTTP/1.1

`,
			givenHTTPResponse: http.Response{Status: "200 OK"},
			givenCeTemplate:   `{ "responseStatus":  {{ .httpresponse.status | quote }} }`,
			whenIncomingEvent: newEventJSON(`{}`),
			thenWantHTTPProtocol: `GET http://localhost:8080/get HTTP/1.1

`,
			thenWantOutgoingEvent: newEventJSON(`{ "responseStatus": "200 OK"}`),
			thenWantErr:           false},

		{name: "Post constant template",
			givenHTTPTemplate: `POST http://localhost:8080/postPerson HTTP/1.1
content-type: application/json

{ "name": "Alex" }`,
			givenHTTPResponse: http.Response{Status: "200 OK", Body: ioutil.NopCloser(bytes.NewBufferString(`{ "name": "Alex", "gender": "male" }`))},
			givenCeTemplate:   `{ "responseStatus":  {{ .httpresponse.status | quote }}, "responseBody": {{ .httpresponse.body | toJson }}  }`,
			whenIncomingEvent: newEventJSON(`{}`),
			thenWantHTTPProtocol: `POST http://localhost:8080/postPerson HTTP/1.1
content-type: application/json

{ "name": "Alex" }`,
			thenWantOutgoingEvent: newEventJSON(`{ "responseStatus": "200 OK", "responseBody": { "gender": "male", "name": "Alex" }}`),
			thenWantErr:           false},

		{name: "Post template",
			givenHTTPTemplate: `POST http://localhost:8080/person/{{ .data.name }} HTTP/1.1
content-type: application/json

{ "gender": {{ .data.gender | quote }} }`,
			givenHTTPResponse: http.Response{Status: "200 OK", Body: ioutil.NopCloser(bytes.NewBufferString(`{ "name": "Alex", "gender": "male" }`))},
			givenCeTemplate:   `{ "person": { "name": {{ .httpresponse.body.name | quote }}, "sex": {{ .httpresponse.body.gender | quote }} } }`,
			whenIncomingEvent: newEventJSON(`{ "name": "Alex", "gender": "male" }`),
			thenWantHTTPProtocol: `POST http://localhost:8080/person/Alex HTTP/1.1
content-type: application/json

{ "gender": "male" }`,
			thenWantOutgoingEvent: newEventJSON(`{ "person": { "sex": "male", "name": "Alex" } }`),
			thenWantErr:           false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			senderCreator := func(protocol string, timeOut time.Duration, debug bool) (HTTPSender, error) {
				if protocol == tt.thenWantHTTPProtocol {
					ms := NewMockHTTPSender(tt.givenHTTPResponse)
					return ms, nil
				}
				t.Errorf("unexpected HTTP-Protocol, actual = '%s', want = '%s'", protocol, tt.thenWantHTTPProtocol)
				return nil, errors.New("unexpected http-protocol")
			}
			ct, err := ceHTTPClientTransformer(senderCreator, tt.givenHTTPTemplate, tt.givenCeTemplate, 5*time.Second, true, true, true)
			if (err != nil) != tt.thenWantErr {
				t.Errorf("cehttpclienttransformer error = %v, wantErr %v", err, tt.thenWantErr)
				return
			}
			outgoingCe, err := ct.TransformEvent(&tt.whenIncomingEvent)
			if (err != nil) != tt.thenWantErr {
				t.Errorf("cehttpclienttransformer.TransformEvent error = %v, wantErr %v", err, tt.thenWantErr)
				return
			}
			cetransformer.CompareEvents(t, "cehttpclienttransformer.TransformEvent", *outgoingCe, tt.thenWantOutgoingEvent)
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
