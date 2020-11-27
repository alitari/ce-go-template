package cehttpclienttransformer

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
	"time"
)

func ReadFromFile(filename string) string {
	content, err := ioutil.ReadFile("../../test/httpRequests/" + filename)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

type HTTPServer struct {
	t                  *testing.T
	expectedMethod     string
	expectedRequestURI string
	expectedBody       string
	srv                *http.Server
}

func setupHTTPServer(t *testing.T, port int, expectedMethod string, expectedRequestURI string, expectedBody string) *HTTPServer {
	server := new(HTTPServer)
	server.expectedMethod = expectedMethod
	server.expectedRequestURI = expectedRequestURI
	server.expectedBody = expectedBody
	server.t = t
	mux := http.NewServeMux()
	mux.HandleFunc("/", server.assertRequest)
	server.srv = &http.Server{Addr: fmt.Sprintf(":%v", port), Handler: mux}

	go func() {
		if err := server.srv.ListenAndServe(); err != nil {
			log.Printf("sink.listenAndServe: %v", err)
		}
	}()
	return server
}

func (s *HTTPServer) ShutDown() error {
	if err := s.srv.Shutdown(context.TODO()); err != nil {
		return err
	}
	return nil
}

func (s *HTTPServer) assertRequest(w http.ResponseWriter, r *http.Request) {
	log.Printf("Request received, %v!", r)

	if s.expectedMethod != r.Method {
		s.t.Errorf("HTTPServer method failure: actual: %s, but expected %s", r.Method, s.expectedMethod)
	}

	if s.expectedRequestURI != r.RequestURI {
		s.t.Errorf("HTTPServer RequestURI failure: actual: '%s', but expected '%s'", r.RequestURI, s.expectedRequestURI)
	}

	actualbody, err := ioutil.ReadAll(r.Body)
	if err != nil {
		s.t.Errorf("HTTPServer error reading request body : %v", err)
	}

	actualbodyStr := string(actualbody)
	if s.expectedBody != actualbodyStr {
		s.t.Errorf("HTTPServer request body not equal: actual = '%s', want '%s'", actualbodyStr, s.expectedBody)
	}

}

func TestHTTPProtocolSender_Send(t *testing.T) {
	type args struct {
		protocol string
	}
	tests := []struct {
		name                   string
		givenTimout            time.Duration
		whenHTTPProtocol       string
		thenWantErr            bool
		thenWantRequestMethod  string
		thenWantRequestURI     string
		thenWantRequestBody    string
		thenWantResponseStatus string
	}{
		{
			name:                   "GetSimple",
			givenTimout:            5 * time.Second,
			whenHTTPProtocol:       ReadFromFile("localTestGetSimple.http"),
			thenWantErr:            false,
			thenWantRequestMethod:  "GET",
			thenWantRequestURI:     "/get",
			thenWantRequestBody:    "",
			thenWantResponseStatus: "200 OK"},
		{
			name:                   "PostSimple",
			givenTimout:            5 * time.Second,
			whenHTTPProtocol:       ReadFromFile("localTestPostSimple.http"),
			thenWantErr:            false,
			thenWantRequestMethod:  "POST",
			thenWantRequestURI:     "/postPerson",
			thenWantRequestBody:    `{ "name": "Alex" }`,
			thenWantResponseStatus: "200 OK"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := setupHTTPServer(t, 8080, tt.thenWantRequestMethod, tt.thenWantRequestURI, tt.thenWantRequestBody)
			time.Sleep(100 * time.Millisecond)
			sender, _ := NewHTTPProtocolSender(tt.whenHTTPProtocol, tt.givenTimout, true)
			response, err := sender.Send()
			if (err != nil) != tt.thenWantErr {
				t.Errorf("HTTPProtocolSender.Send() error = %v, wantErr %v", err, tt.thenWantErr)
			} else {
				if response.Status != tt.thenWantResponseStatus {
					t.Errorf("HTTPProtocolSender response status failure: actual: %s, but want %s", response.Status, tt.thenWantResponseStatus)
				}
			}
			time.Sleep(100 * time.Millisecond)
			server.ShutDown()
		})
	}
}
