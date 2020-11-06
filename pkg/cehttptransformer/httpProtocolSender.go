package cehttptransformer

import (
	"bufio"
	"log"
	"net/http"
	"strings"
	"time"
)

// HTTPProtocolSender bla
type HTTPProtocolSender struct {
	request *http.Request
	client  *http.Client
}

func NewHTTPProtocolSender(protocol string, timeout time.Duration) *HTTPProtocolSender {
	hps := new(HTTPProtocolSender)
	if hps.client == nil {
		hps.client = &http.Client{
			Timeout: timeout,
		}
	}
	log.Printf("HTTP Request Protocol:\n%s\n", protocol)
	buf := bufio.NewReader(strings.NewReader(protocol))
	request, err := http.ReadRequest(buf)
	if err != nil {
		log.Printf("error: %v", err)
		return nil
	}
	hps.request = request
	return hps
}

// Send bla
func (hps *HTTPProtocolSender) Send() (*http.Response, error) {
	hps.request.RequestURI = ""
	response, err := hps.client.Do(hps.request)
	if err != nil {
		return nil, err
	}
	return response, nil
}
