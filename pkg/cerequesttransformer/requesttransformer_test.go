package cerequesttransformer

import (
	"log"
	"net/http"
	"strings"
	"testing"

	"github.com/alitari/ce-go-template/pkg/cetransformer"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

func TestRequestTransformer(t *testing.T) {

	tests := []struct {
		name            string
		givenCeTemplate string
		givenCeSource   string
		givenCeType     string
		whenRequest     *http.Request
		thenWantError   bool
		thenWantEvent   cloudevents.Event
	}{
		{
			name:            "http request",
			givenCeTemplate: `{ "method": {{ .method | quote }} , "header": {{ .header | toJson }}, "url": {{ .url | toJson }}, "body": {{ .body | toJson }} }`,
			givenCeSource:   "mysource",
			givenCeType:     "mytype",
			whenRequest:     NewReq("GET", map[string][]string{"Content-Type": {"application/json"}}, "http://foo.bar:8080/mypath", `{ "name": "Alex" }`),
			thenWantError:   false,
			thenWantEvent: cetransformer.NewEventWithJSONStringData(`
{
	 "method": "GET",
	 "url": {
              "hostname": "foo.bar",
              "path": "/mypath",
              "query": {},
              "scheme": "http"
			},
	 "header": {
              "Content-Type": [
                "application/json"
              ]
			},
	 "body": {
         "name": "Alex"
     }
}
`, "mysource", "mytype")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rt, err := NewRequestTransformer(tt.givenCeTemplate, tt.givenCeType, tt.givenCeSource, true)
			if (err != nil) != tt.thenWantError {
				t.Errorf("can't create requesttransformer error = %v, wantErr %v", err, tt.thenWantError)
				return
			}
			actualEvent, err := rt.CreateEvent(*tt.whenRequest)
			if (err != nil) != tt.thenWantError {
				t.Errorf("cehttpclienttransformer.TransformEvent error = %v, wantErr %v", err, tt.thenWantError)
				return
			}
			cetransformer.CompareEvents(t, "RequestTransformer.CreateEvent", *actualEvent, tt.thenWantEvent)

		})
	}
}

func NewReq(method string, header http.Header, url, body string) *http.Request {
	req, err := http.NewRequest(method, url, strings.NewReader(body))
	req.Header = header
	if err != nil {
		log.Fatalf("Can't create request error = %v", err)
		return nil
	}
	return req
}
