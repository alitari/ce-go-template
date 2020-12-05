package cehttpclienttransformer

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/alitari/ce-go-template/pkg/cetransformer"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

const defaultEventJSONMetaData = `"datacontenttype":"application/json","id":"id","source":"source","specversion":"1.0","type":"type"}`

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

func TestTransformEvent(t *testing.T) {
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
			whenIncomingEvent: cetransformer.NewEventWithJSONStringData(`{}`),
			thenWantHTTPProtocol: `GET http://localhost:8080/get HTTP/1.1

`,
			thenWantOutgoingEvent: cetransformer.NewEventWithJSONStringData(`{ "responseStatus": "200 OK"}`),
			thenWantErr:           false},

		{name: "Post constant template",
			givenHTTPTemplate: `POST http://localhost:8080/postPerson HTTP/1.1
content-type: application/json

{ "name": "Alex" }`,
			givenHTTPResponse: http.Response{Status: "200 OK", Body: ioutil.NopCloser(bytes.NewBufferString(`{ "name": "Alex", "gender": "male" }`))},
			givenCeTemplate:   `{ "responseStatus":  {{ .httpresponse.status | quote }}, "responseBody": {{ .httpresponse.body | toJson }}  }`,
			whenIncomingEvent: cetransformer.NewEventWithJSONStringData(`{}`),
			thenWantHTTPProtocol: `POST http://localhost:8080/postPerson HTTP/1.1
content-type: application/json

{ "name": "Alex" }`,
			thenWantOutgoingEvent: cetransformer.NewEventWithJSONStringData(`{ "responseStatus": "200 OK", "responseBody": { "gender": "male", "name": "Alex" }}`),
			thenWantErr:           false},

		{name: "Post template",
			givenHTTPTemplate: `POST http://localhost:8080/person/{{ .data.name }} HTTP/1.1
content-type: application/json

{ "gender": {{ .data.gender | quote }} }`,
			givenHTTPResponse: http.Response{Status: "200 OK", Body: ioutil.NopCloser(bytes.NewBufferString(`{ "name": "Alex", "gender": "male" }`))},
			givenCeTemplate:   `{ "person": { "name": {{ .httpresponse.body.name | quote }}, "sex": {{ .httpresponse.body.gender | quote }} } }`,
			whenIncomingEvent: cetransformer.NewEventWithJSONStringData(`{ "name": "Alex", "gender": "male" }`),
			thenWantHTTPProtocol: `POST http://localhost:8080/person/Alex HTTP/1.1
content-type: application/json

{ "gender": "male" }`,
			thenWantOutgoingEvent: cetransformer.NewEventWithJSONStringData(`{ "person": { "sex": "male", "name": "Alex" } }`),
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
			ct, err := ceHTTPClientTransformer(senderCreator, tt.givenHTTPTemplate, tt.givenCeTemplate, 5*time.Second, true, true)
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

func TestPredicateEvent(t *testing.T) {
	tests := []struct {
		name                 string
		givenHTTPTemplate    string
		givenCeTemplate      string
		givenHTTPResponse    http.Response
		whenIncomingEvent    cloudevents.Event
		thenWantHTTPProtocol string
		thenWantPredicate    bool
		thenWantErr          bool
	}{
		{name: "Get constant template true",
			givenHTTPTemplate: `GET http://localhost:8080/get HTTP/1.1

`,
			givenHTTPResponse: http.Response{Status: "200 OK"},
			givenCeTemplate:   `{{ eq .httpresponse.status "200 OK" | toString }}`,
			whenIncomingEvent: cetransformer.NewEventWithJSONStringData(`{}`),
			thenWantHTTPProtocol: `GET http://localhost:8080/get HTTP/1.1

`,
			thenWantPredicate: true,
			thenWantErr:       false},
		{name: "Get constant template false",
			givenHTTPTemplate: `GET http://localhost:8080/get HTTP/1.1

`,
			givenHTTPResponse: http.Response{Status: "500 Internal server error"},
			givenCeTemplate:   `{{ eq .httpresponse.status "200 OK" | toString }}`,
			whenIncomingEvent: cetransformer.NewEventWithJSONStringData(`{}`),
			thenWantHTTPProtocol: `GET http://localhost:8080/get HTTP/1.1

`,
			thenWantPredicate: false,
			thenWantErr:       false},
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
			ct, err := ceHTTPClientTransformer(senderCreator, tt.givenHTTPTemplate, tt.givenCeTemplate, 5*time.Second, true, true)
			if (err != nil) != tt.thenWantErr {
				t.Errorf("cehttpclienttransformer error = %v, wantErr %v", err, tt.thenWantErr)
				return
			}
			predicate, err := ct.PredicateEvent(&tt.whenIncomingEvent)
			if (err != nil) != tt.thenWantErr {
				t.Errorf("cehttpclienttransformer.TransformEvent error = %v, wantErr %v", err, tt.thenWantErr)
				return
			}
			if predicate != tt.thenWantPredicate {
				t.Errorf("cehttpclienttransformer.Predicate is not equal: actual = %v, want %v", predicate, tt.thenWantPredicate)
			}
		})
	}
}

func TestResponseToMap(t *testing.T) {
	tests := []struct {
		name          string
		givenJSONBody bool
		whenResponse  *http.Response
		thenWant      map[string]interface{}
		thenWantErr   bool
	}{
		{name: "Simple",
			givenJSONBody: true,
			whenResponse:  &http.Response{Header: http.Header{"Content-Type": {"application/json"}}, Status: "200 OK", StatusCode: 200, Body: ioutil.NopCloser(bytes.NewBufferString(`{ "name": "Alex", "gender": "male" }`))},
			thenWant:      map[string]interface{}{"header": http.Header{"Content-Type": {"application/json"}}, "body": map[string]interface{}{"name": "Alex", "gender": "male"}, "status": "200 OK", "statusCode": 200}, thenWantErr: false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ResponseToMap(tt.whenResponse, tt.givenJSONBody)
			if (err != nil) != tt.thenWantErr {
				t.Errorf("ResponseToMap() error = %v, wantErr %v", err, tt.thenWantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.thenWant) {
				t.Errorf("ResponseToMap() = %v, want %v", got, tt.thenWant)
			}
		})
	}
}
