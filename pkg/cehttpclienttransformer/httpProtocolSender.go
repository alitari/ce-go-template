package cehttpclienttransformer

import (
	"bufio"
	"io/ioutil"
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

// NewHTTPProtocolSender bla
func NewHTTPProtocolSender(protocol string, timeout time.Duration, debug bool) (*HTTPProtocolSender, error) {
	hps := new(HTTPProtocolSender)
	if hps.client == nil {
		hps.client = &http.Client{
			Timeout: timeout,
		}
	}

	if debug {
		log.Printf("HTTP Request String:\n%s\n", protocol)
	}
	buf := bufio.NewReader(strings.NewReader(protocol))
	request, err := http.ReadRequest(buf)
	if err != nil {
		return nil, err
	}
	reqStr := strings.Split(protocol, "\n\n")
	if len(reqStr) > 1 {
		if debug {
			log.Printf("HTTP Request Body:\n%s\n", reqStr[1])
		}
		request.Body = ioutil.NopCloser(strings.NewReader(reqStr[1]))
	}
	hps.request = request
	return hps, nil
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
