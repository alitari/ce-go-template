package cehttpserver

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/alitari/ce-go-template/pkg/cehandler"
	"github.com/alitari/ce-go-template/pkg/cerequesttransformer"
)

var sinkPort = 8099

func createProducerHandler(cetemplate string) (*cehandler.CeProducerHandler, error) {
	producer, err := cerequesttransformer.NewRequestTransformer(cetemplate, "type", "source", true)
	if err != nil {
		return nil, err
	}
	return cehandler.NewProducerHandler(producer, fmt.Sprintf("http://localhost:%v/", sinkPort), 5*time.Second, true), nil
}

type SinkServer struct {
	t              *testing.T
	expectedSource string
	expectedType   string
	expectedBody   string
	srv            *http.Server
}

func setupSink(t *testing.T, expectedSource string, expectedType string, expectedBody string) *SinkServer {
	sink := new(SinkServer)
	sink.expectedSource = expectedSource
	sink.expectedType = expectedType
	sink.expectedBody = expectedBody
	sink.t = t
	mux := http.NewServeMux()
	mux.HandleFunc("/", sink.assertResultMessage)
	sink.srv = &http.Server{Addr: fmt.Sprintf(":%v", sinkPort), Handler: mux}

	go func() {
		if err := sink.srv.ListenAndServe(); err != nil {
			log.Printf("sink.listenAndServe: %v", err)
		}
	}()
	return sink
}

func (s *SinkServer) ShutDown() error {
	if err := s.srv.Shutdown(context.TODO()); err != nil {
		return err
	}
	return nil
}

func (s *SinkServer) assertResultMessage(w http.ResponseWriter, r *http.Request) {
	log.Printf("Result received, %v!", r)

	actualSource := r.Header["Ce-Source"][0]
	actualType := r.Header["Ce-Type"][0]

	if s.expectedSource != actualSource {
		s.t.Errorf("cehttpservertransformer sink request source header not equal: actual = '%s', want '%s'", actualSource, s.expectedSource)
	}

	if s.expectedType != actualType {
		s.t.Errorf("cehttpservertransformer sink request type header not equal: actual = '%s', want '%s'", actualType, s.expectedType)
	}

	actualbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.t.Errorf("cehttpservertransformer sink error reading request body : %v", err)
	}

	actualbodyStr := string(actualbody)
	if s.expectedBody != actualbodyStr {
		s.t.Errorf("cehttpservertransformer sink request body not equal: actual = '%s', want '%s'", actualbodyStr, s.expectedBody)
	}

}

// ("GET", fmt.Sprintf("http://localhost:%v%s", tt.port, tt.path), nil)
func NewRequest(method string, port int, path string, body string) *http.Request {
	req, err := http.NewRequest("GET", fmt.Sprintf("http://localhost:%v%s", port, path), strings.NewReader(body))
	if err != nil {
		log.Fatalf("Can't create request error = %v", err)
		return nil
	}
	return req
}

func Testcehttpservertransformer(t *testing.T) {
	tests := []struct {
		name                  string
		givenServerPort       int
		givenServerPath       string
		givenServerMethod     string
		givenServerCetemplate string
		whenRequest           *http.Request
		thenWantErr           bool
		thenWantSource        string
		thenWantType          string
		thenWantBody          string
	}{
		{
			name:            "constant",
			givenServerPort: 8080, givenServerPath: "/path", givenServerMethod: "GET", givenServerCetemplate: `
{ "data": "",
	"datacontenttype":"application/json",
	"id": "{{ uuidv4 }}",
	"source": "testsource",
	"specversion": "1.0",
	"type": "type" 
}`,
			whenRequest:    NewRequest("GET", 8080, "/path", "doesn't matter"),
			thenWantSource: "testsource", thenWantType: "type", thenWantBody: `""`},
		{
			name:            "simple",
			givenServerPort: 8080, givenServerPath: "/path", givenServerMethod: "GET", givenServerCetemplate: `
{ "data": "",
	"datacontenttype":"application/json",
	"id": "{{ uuidv4 }}",
	"source": "testsource",
	"specversion": "1.0",
	"type": "type" 
}`,
			whenRequest:    NewRequest("GET", 8080, "/path", "doesn't matter"),
			thenWantSource: "testsource", thenWantType: "type", thenWantBody: `""`},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sink := setupSink(t, tt.thenWantSource, tt.thenWantType, tt.thenWantBody)
			producerHandler, err := createProducerHandler(tt.givenServerCetemplate)
			if (err != nil) != tt.thenWantErr {
				t.Errorf("cehttpservertransformer error = %v, wantErr %v", err, tt.thenWantErr)
				return
			}
			cehttpservertransformer := NewCeHTTPServer(tt.givenServerPort, tt.givenServerPath, tt.givenServerMethod, true, producerHandler)
			resp, err := http.DefaultClient.Do(tt.whenRequest)
			if (err != nil) != tt.thenWantErr {
				t.Errorf("cehttpservertransformer error = %v, wantErr %v", err, tt.thenWantErr)
			}
			if resp.StatusCode != 200 {
				t.Errorf("cehttpservertransformer expect 200 response, but is %v", resp.StatusCode)
			}
			time.Sleep(1 * time.Second)
			cehttpservertransformer.ShutDown()
			sink.ShutDown()
			time.Sleep(1 * time.Second)
		})
	}
}
