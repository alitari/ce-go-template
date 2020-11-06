package cehttptransformer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"
)

func readFromFile(filename string) string {
	content, err := ioutil.ReadFile("../../test/httpRequests/" + filename)
	if err != nil {
		log.Fatal(err)
	}
	return string(content)
}

func TestHTTPProtocolSender_Parse_Send_Map(t *testing.T) {
	type args struct {
		protocol string
	}
	tests := []struct {
		name       string
		hps        *HTTPProtocolSender
		wantErr    bool
		jsonBody   bool
		assertFunc func(response map[string]interface{}) string
	}{
		{name: "SimpleGet", hps: NewHTTPProtocolSender(readFromFile("httpbinGet.http"), 5*time.Second), wantErr: false,
			assertFunc: func(response map[string]interface{}) string {
				if response["statusCode"] != 200 {
					return "statusCode not ok"
				}
				header := response["header"].(http.Header)
				ct := header["Content-Type"]
				if ct[0] != "application/json" {
					return "expect Content-Type: application/json"
				}
				bodyData := response["body"]
				bodyBytes, err := json.Marshal(bodyData)
				if err != nil {
					return fmt.Sprintf("can't marshall body: %v", err)
				}
				bodyStr := string(bodyBytes)
				if !strings.Contains(bodyStr, "httpbin.org") {
					return fmt.Sprintf("unexpected body: %v", bodyStr)
				}
				return ""
			}},
		{name: "SimplePost", hps: NewHTTPProtocolSender(readFromFile("httpbinPost.http"), 5*time.Second), wantErr: false,
			assertFunc: func(response map[string]interface{}) string {
				return ""
			}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			response, err := tt.hps.Send()
			if (err != nil) != tt.wantErr {
				t.Errorf("HTTPProtocolSender.Send() error = %v, wantErr %v", err, tt.wantErr)
			} else {
				responseMap, err := ResponseToMap(response, tt.jsonBody)
				if (err != nil) != tt.wantErr {
					t.Errorf("HTTPProtocolSender.Map() error = %v, wantErr %v", err, tt.wantErr)
				} else {
					if msg := tt.assertFunc(responseMap); len(msg) > 0 {
						t.Errorf("HTTPProtocolSender response assert failure: %s", msg)
					}
					log.Printf("Response: %v", response)
				}
			}
		})
	}
}
