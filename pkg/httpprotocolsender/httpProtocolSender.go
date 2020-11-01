package httpprotocolsender

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
)

// HTTPProtocolSender bla
type HTTPProtocolSender struct {
	request *http.Request
	client  *http.Client
	Timeout time.Duration
}

// Parse bla
func (hps *HTTPProtocolSender) Parse(protocol string) error {
	if hps.client == nil {
		hps.client = &http.Client{
			Timeout: hps.Timeout,
		}
	}
	log.Printf("HTTP Request Protocol:\n%s\n", protocol)
	buf := bufio.NewReader(strings.NewReader(protocol))
	request, err := http.ReadRequest(buf)
	if err != nil {
		return err
	}
	hps.request = request
	return nil
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

// ResponseToMap transform response into map
func (hps *HTTPProtocolSender) ResponseToMap(response *http.Response, jsonBody bool) (map[string]interface{}, error) {
	responseMap := map[string]interface{}{}
	if response != nil {
		responseMap["header"] = response.Header
		responseMap["statusCode"] = response.StatusCode
		b := new(bytes.Buffer)
		io.Copy(b, response.Body)
		response.Body.Close()
		if jsonBody {
			bodyData := map[string]interface{}{}
			if err := json.Unmarshal(b.Bytes(), &bodyData); err != nil {
				return nil, err
			}
			responseMap["body"] = bodyData
		} else {
			responseMap["body"] = b.String()
		}
	}
	return responseMap, nil
}
